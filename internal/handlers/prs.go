package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"reviewer-service/internal/db"
	"reviewer-service/internal/services"

	"github.com/gorilla/mux"
)

func RegisterPRRoutes(r *mux.Router) {
	r.HandleFunc("/pullRequest/create", createPR).Methods("POST")
	r.HandleFunc("/pullRequest/merge", mergePR).Methods("POST")
	r.HandleFunc("/pullRequest/reassign", reassignPR).Methods("POST")
}

// createPR expects { pull_request_id, pull_request_name, author_id }
func createPR(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, ErrCodeNotFound, err.Error())
		return
	}
	if body.PullRequestID == "" || body.PullRequestName == "" ||
		body.AuthorID == "" {
		writeErr(w, 400, ErrCodeNotFound,
			"pull_request_id, pull_request_name and author_id required")
		return
	}
	// check author exists
	var a string
	if err := db.DB.QueryRow(`SELECT id FROM users WHERE id=$1`,
		body.AuthorID).Scan(&a); err != nil {
		if err == sql.ErrNoRows {
			writeErr(w, 404, ErrCodeNotFound, "author not found")
			return
		}
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	// find author team (optional)
	var teamName sql.NullString
	_ = db.DB.QueryRow(`SELECT team_name FROM team_users WHERE user_id=$1
LIMIT 1`, body.AuthorID).Scan(&teamName)
	// try create PR; if exists -> PR_EXISTS
	_, err := db.DB.Exec(`INSERT INTO prs(id,title,author_id,team_name)
VALUES($1,$2,$3,$4)`, body.PullRequestID, body.PullRequestName,
		body.AuthorID, teamName)
	if err != nil {
		// conflict
		writeErr(w, 409, ErrCodePRExists, "PR id already exists")
		return
	}
	// assign up to two reviewers if team present
	if teamName.Valid {
		if err := services.AssignUpToTwoReviewers(body.PullRequestID,
			body.AuthorID, teamName.String); err != nil {
			writeErr(w, 500, ErrCodeNotFound, err.Error())
			return
		}
	}
	pr, err := loadPRForAPI(body.PullRequestID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(struct {
		PR interface{} `json:"pr"`
	}{PR: pr})
}
func mergePR(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PullRequestID string `json:"pull_request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, ErrCodeNotFound, err.Error())
		return
	}
	if body.PullRequestID == "" {
		writeErr(w, 400, ErrCodeNotFound, "pull_request_id required")
		return
	}
	res, err := db.DB.Exec(`UPDATE prs SET status='MERGED', merged_at=now()
WHERE id=$1 AND status!='MERGED'`, body.PullRequestID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		// could be not found or already merged; check existence
		var exists string
		if err := db.DB.QueryRow(`SELECT id FROM prs WHERE id=$1`,
			body.PullRequestID).Scan(&exists); err != nil {
			writeErr(w, 404, ErrCodeNotFound, "pr not found")
			return
		}
		// existed but not updated -> already merged
	}
	pr, err := loadPRForAPI(body.PullRequestID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	json.NewEncoder(w).Encode(struct {
		PR interface{} `json:"pr"`
	}{PR: pr})
}

// reassignPR: { pull_request_id, old_reviewer_id }
func reassignPR(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, ErrCodeNotFound, err.Error())
		return
	}
	if body.PullRequestID == "" || body.OldReviewerID == "" {
		writeErr(w, 400, ErrCodeNotFound, "pull_request_id and old_reviewer_id required")
		return
	}
	// Check PR exists and not merged
	var status string
	if err := db.DB.QueryRow(`SELECT status FROM prs WHERE id=$1`,
		body.PullRequestID).Scan(&status); err != nil {
		writeErr(w, 404, ErrCodeNotFound, "pr not found")
		return
	}
	if status == "MERGED" {
		writeErr(w, 409, ErrCodePRMerged, "cannot reassign on merged PR")
		return
	}
	// check old reviewer is assigned
	var dummy string
	if err := db.DB.QueryRow(`SELECT reviewer_id FROM pr_reviewers WHERE
pr_id=$1 AND reviewer_id=$2`, body.PullRequestID,
		body.OldReviewerID).Scan(&dummy); err != nil {
		if err == sql.ErrNoRows {
			writeErr(w, 409, ErrCodeNotAssigned,
				"reviewer is not assigned to this PR")
			return
		}
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	// find team of old reviewer
	var teamName string
	if err := db.DB.QueryRow(`SELECT team_name FROM team_users WHERE
user_id=$1 LIMIT 1`, body.OldReviewerID).Scan(&teamName); err != nil {
		writeErr(w, 404, ErrCodeNotFound, "reviewer has no team")
		return
	}
	// collect current assigned reviewers to exclude
	rows, err := db.DB.Query(`SELECT reviewer_id FROM pr_reviewers WHERE
pr_id=$1`, body.PullRequestID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	defer rows.Close()
	var assigned []string
	for rows.Next() {
		var id string
		_ = rows.Scan(&id)
		assigned = append(assigned, id)
	}
	// pick candidate excluding old reviewer and currently assigned
	candidate, err := services.PickRandomActiveFromTeamExcluding(teamName,
		append(assigned, body.OldReviewerID))
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	if candidate == "" {
		writeErr(w, 409, ErrCodeNoCandidate,
			"no active replacement candidate in team")
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`DELETE FROM pr_reviewers WHERE pr_id=$1 AND
reviewer_id=$2`, body.PullRequestID, body.OldReviewerID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	_, err = tx.Exec(`INSERT INTO pr_reviewers(pr_id, reviewer_id)
VALUES($1,$2)`, body.PullRequestID, candidate)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	pr, err := loadPRForAPI(body.PullRequestID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	json.NewEncoder(w).Encode(struct {
		PR         interface{} `json:"pr"`
		ReplacedBy string      `json:"replaced_by"`
	}{PR: pr, ReplacedBy: candidate})
}

// helper: load PR and form API shape
func loadPRForAPI(id string) (map[string]interface{}, error) {
	var title, author, team, status string
	var createdAt sql.NullTime
	var mergedAt sql.NullTime
	row := db.DB.QueryRow(`SELECT
id,title,author_id,team_name,status,created_at,merged_at FROM prs WHERE
id=$1`, id)
	if err := row.Scan(&id, &title, &author, &team, &status, &createdAt,
		&mergedAt); err != nil {
		return nil, err
	}
	rows, err := db.DB.Query(`SELECT reviewer_id FROM pr_reviewers WHERE
pr_id=$1 ORDER BY assigned_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reviewers []string
	for rows.Next() {
		var rid string
		_ = rows.Scan(&rid)
		reviewers = append(reviewers, rid)
	}
	m := map[string]interface{}{
		"pull_request_id":    id,
		"pull_request_name":  title,
		"author_id":          author,
		"status":             status,
		"assigned_reviewers": reviewers,
	}
	if createdAt.Valid {
		m["createdAt"] = createdAt.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	if mergedAt.Valid {
		m["mergedAt"] = mergedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
	}
	return m, nil
}
