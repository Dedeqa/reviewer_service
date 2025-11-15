package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"reviewer-service/internal/db"
	"reviewer-service/internal/models"
	"reviewer-service/internal/services"
)

func RegisterPRRoutes(r *mux.Router) {
	r.HandleFunc("/prs", createPR).Methods("POST")
	r.HandleFunc("/prs/{id}", getPR).Methods("GET")
	r.HandleFunc("/prs/{id}/merge", mergePR).Methods("POST")
	r.HandleFunc("/prs/{id}/reassign", reassignPR).Methods("POST")
}

func createPR(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Title    string `json:"title"`
		AuthorID string `json:"authorId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeErr(w, 400, err)
		return
	}

	if p.Title == "" || p.AuthorID == "" {
		writeErr(w, 400, errors.New("title and authorId required"))
		return
	}

	// author team
	var teamName string
	err := db.DB.QueryRow(
		`SELECT team_name FROM team_users WHERE user_id=$1 LIMIT 1`,
		p.AuthorID,
	).Scan(&teamName)
	if err != nil {
		teamName = ""
	}

	// create PR
	var prID string
	err = db.DB.QueryRow(`
        INSERT INTO prs(title,author_id,team_name)
        VALUES($1,$2,$3) RETURNING id
    `, p.Title, p.AuthorID, teamName).Scan(&prID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	if teamName != "" {
		if err := services.AssignUpToTwoReviewers(prID, p.AuthorID, teamName); err != nil {
			writeErr(w, 500, err)
			return
		}
	}

	pr, err := services.LoadPR(prID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(pr)
}

func getPR(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	pr, err := services.LoadPR(id)
	if err != nil {
		writeErr(w, 404, err)
		return
	}

	json.NewEncoder(w).Encode(pr)
}

func mergePR(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	_, _ = db.DB.Exec(`
        UPDATE prs SET status='MERGED'
        WHERE id=$1 AND status!='MERGED'
    `, id)

	pr, err := services.LoadPR(id)
	if err != nil {
		writeErr(w, 404, err)
		return
	}

	json.NewEncoder(w).Encode(pr)
}

func reassignPR(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var body struct {
		ReviewerID string `json:"reviewerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, err)
		return
	}
	if body.ReviewerID == "" {
		writeErr(w, 400, errors.New("reviewerId required"))
		return
	}

	// check is merged
	var status string
	if err := db.DB.QueryRow(`SELECT status FROM prs WHERE id=$1`, id).Scan(&status); err != nil {
		writeErr(w, 404, errors.New("pr not found"))
		return
	}
	if status == string(models.StatusMerged) {
		writeErr(w, 400, errors.New("cannot reassign reviewers for merged PR"))
		return
	}

	var team string
	if err := db.DB.QueryRow(
		`SELECT team_name FROM team_users WHERE user_id=$1 LIMIT 1`,
		body.ReviewerID,
	).Scan(&team); err != nil {
		writeErr(w, 400, errors.New("reviewer has no team"))
		return
	}

	candidateID, err := services.PickRandomActiveFromTeamExcluding(team, []string{body.ReviewerID})
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if candidateID == "" {
		writeErr(w, 400, errors.New("no available candidate to reassign"))
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`DELETE FROM pr_reviewers WHERE pr_id=$1 AND reviewer_id=$2`,
		id, body.ReviewerID,
	)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	_, err = tx.Exec(
		`INSERT INTO pr_reviewers(pr_id,reviewer_id) VALUES($1,$2)`,
		id, candidateID,
	)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	if err := tx.Commit(); err != nil {
		writeErr(w, 500, err)
		return
	}

	pr, err := services.LoadPR(id)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	json.NewEncoder(w).Encode(pr)
}
