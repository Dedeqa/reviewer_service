package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"reviewer-service/internal/db"
	"reviewer-service/internal/models"
)

func RegisterTeamRoutes(r *mux.Router) {
	r.HandleFunc("/teams", createTeam).Methods("POST")
	r.HandleFunc("/teams/{name}", getTeam).Methods("GET")
	r.HandleFunc("/teams/{name}/users", addUserToTeam).Methods("POST")
}

func createTeam(w http.ResponseWriter, r *http.Request) {
	var t models.Team
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeErr(w, 400, err)
		return
	}
	if t.Name == "" {
		writeErr(w, 400, errors.New("team name required"))
		return
	}

	_, err := db.DB.Exec(`INSERT INTO teams(name) VALUES($1) ON CONFLICT DO NOTHING`, t.Name)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func getTeam(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var t models.Team
	row := db.DB.QueryRow(`SELECT name FROM teams WHERE name=$1`, name)
	if err := row.Scan(&t.Name); err != nil {
		writeErr(w, 404, errors.New("team not found"))
		return
	}

	json.NewEncoder(w).Encode(t)
}

func addUserToTeam(w http.ResponseWriter, r *http.Request) {
	teamName := mux.Vars(r)["name"]

	var body struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, err)
		return
	}

	if body.UserID == "" {
		writeErr(w, 400, errors.New("user_id required"))
		return
	}

	// Проверяем, существует ли команда
	var exists string
	if err := db.DB.QueryRow(`SELECT name FROM teams WHERE name=$1`, teamName).Scan(&exists); err != nil {
		writeErr(w, 404, errors.New("team not found"))
		return
	}

	// Проверяем, существует ли пользователь
	if err := db.DB.QueryRow(`SELECT id FROM users WHERE id=$1`, body.UserID).Scan(&exists); err != nil {
		writeErr(w, 404, errors.New("user not found"))
		return
	}

	// Вставляем в team_users (ON CONFLICT DO NOTHING чтобы не дублировать)
	_, err := db.DB.Exec(`
        INSERT INTO team_users(team_name, user_id)
        VALUES($1, $2)
        ON CONFLICT DO NOTHING
    `, teamName, body.UserID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"ok"}`))
}
