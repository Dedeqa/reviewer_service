package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"reviewer-service/internal/db"
	"reviewer-service/internal/models"

	"github.com/gorilla/mux"
)

func RegisterUserRoutes(r *mux.Router) {
	r.HandleFunc("/users/setIsActive", setUserIsActive).Methods("POST")
	r.HandleFunc("/users/getReview", listAssignedPRsHandler).Methods("GET")
}

// setIsActive: body { user_id, is_active }
func setUserIsActive(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, ErrCodeNotFound, err.Error())
		return
	}
	if body.UserID == "" {
		writeErr(w, 400, ErrCodeNotFound, "user_id required")
		return
	}
	res, err := db.DB.Exec(`UPDATE users SET is_active=$1 WHERE id=$2`,
		body.IsActive, body.UserID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		writeErr(w, 404, ErrCodeNotFound, "user not found")
		return
	}
	// try to find a team name for response (optional)
	var teamName sql.NullString
	_ = db.DB.QueryRow(`SELECT team_name FROM team_users WHERE user_id=$1
LIMIT 1`, body.UserID).Scan(&teamName)
	user := models.User{
		ID:       body.UserID,
		Name:     "", // name not supplied by this endpoint
		IsActive: body.IsActive,
	}
	resp := struct {
		User models.User `json:"user"`
		Team string      `json:"team_name,omitempty"`
	}{User: user}
	if teamName.Valid {
		resp.Team = teamName.String
	}
	w.Header().Set("Content-Type", "application/json")
	err2 := json.NewEncoder(w).Encode(resp)
	if err2 != nil {
		return
	}
}

// listAssignedPRsHandler -> GET /users/getReview?user_id=...
func listAssignedPRsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeErr(w, 400, ErrCodeNotFound, "user_id required")
		return
	}
	rows, err := db.DB.Query(`
 	SELECT p.id, p.title, p.author_id, p.status
 	FROM prs p
 	JOIN pr_reviewers prr ON p.id = prr.pr_id
 	WHERE prr.reviewer_id=$1
 	`, userID)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Failed to rollback transaction: %v", err)
		}
	}(rows)
	type prShort struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	}
	var prs []prShort
	for rows.Next() {
		var id, title, author, status string
		if err := rows.Scan(&id, &title, &author, &status); err != nil {
			writeErr(w, 500, ErrCodeNotFound, err.Error())
			return
		}
		prs = append(prs, prShort{PullRequestID: id, PullRequestName: title,
			AuthorID: author, Status: status})
	}
	resp := struct {
		UserID       string    `json:"user_id"`
		PullRequests []prShort `json:"pull_requests"`
	}{UserID: userID, PullRequests: prs}
	w.Header().Set("Content-Type", "application/json")
	err2 := json.NewEncoder(w).Encode(resp)
	if err2 != nil {
		return
	}
}
