package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"reviewer-service/internal/db"
	"reviewer-service/internal/models"
)

func RegisterUserRoutes(r *mux.Router) {
	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users/{id}", updateUser).Methods("PATCH")
	r.HandleFunc("/users/{id}/assigned-prs", listAssignedPRs).Methods("GET")
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		writeErr(w, 400, err)
		return
	}

	if u.Name == "" {
		writeErr(w, 400, errors.New("name required"))
		return
	}

	if u.ID == "" {
		err := db.DB.QueryRow(`
            INSERT INTO users(name,is_active) 
            VALUES($1,$2) 
            RETURNING id
        `, u.Name, u.IsActive).Scan(&u.ID)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
	} else {
		_, err := db.DB.Exec(`
            INSERT INTO users(id,name,is_active)
            VALUES($1,$2,$3)
        `, u.ID, u.Name, u.IsActive)
		if err != nil {
			writeErr(w, 500, err)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var u models.User
	row := db.DB.QueryRow(`SELECT id,name,is_active FROM users WHERE id=$1`, id)

	if err := row.Scan(&u.ID, &u.Name, &u.IsActive); err != nil {
		writeErr(w, 404, errors.New("user not found"))
		return
	}

	json.NewEncoder(w).Encode(u)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var patch struct {
		IsActive *bool `json:"isActive"`
	}

	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		writeErr(w, 400, err)
		return
	}

	if patch.IsActive == nil {
		writeErr(w, 400, errors.New("isActive required"))
		return
	}

	_, err := db.DB.Exec(
		`UPDATE users SET is_active=$1 WHERE id=$2`,
		*patch.IsActive, id,
	)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func listAssignedPRs(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	rows, err := db.DB.Query(`
        SELECT p.id, p.title, p.author_id, p.team_name, p.status
        FROM prs p
        JOIN pr_reviewers prr ON p.id = prr.pr_id
        WHERE prr.reviewer_id=$1
    `, id)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	defer rows.Close()

	var result []models.PR

	for rows.Next() {
		var p models.PR
		var status string

		if err := rows.Scan(&p.ID, &p.Title, &p.AuthorID, &p.TeamName, &status); err != nil {
			writeErr(w, 500, err)
			return
		}

		p.Status = models.PRStatus(status)
		result = append(result, p)
	}

	json.NewEncoder(w).Encode(result)
}
