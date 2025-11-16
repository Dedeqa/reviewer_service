package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"reviewer-service/internal/db"

	"github.com/gorilla/mux"
)

// structures matching OpenAPI Team/TeamMember
type teamMemberReq struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}
type teamReq struct {
	TeamName string          `json:"team_name"`
	Members  []teamMemberReq `json:"members"`
}
type teamResp struct {
	Team teamReq `json:"team"`
}

func RegisterTeamRoutes(r *mux.Router) {
	r.HandleFunc("/team/add", createTeamWithMembers).Methods("POST")
	r.HandleFunc("/team/get", getTeamWithMembers).Methods("GET")
}

// createTeamWithMembers: creates team and upserts users and memberships
func createTeamWithMembers(w http.ResponseWriter, r *http.Request) {
	var t teamReq
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeErr(w, 400, ErrCodeNotFound, err.Error())
		return
	}
	if t.TeamName == "" {
		writeErr(w, 400, ErrCodeNotFound, "team_name required")
		return
	}
	tx, err := db.DB.Begin()
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO teams(name) VALUES($1) ON CONFLICT DO
NOTHING`, t.TeamName)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	// upsert users and insert into team_users
	for _, m := range t.Members {
		// upsert user row
		_, err = tx.Exec(`INSERT INTO users(id,name,is_active)
VALUES($1,$2,$3) ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name,
is_active=EXCLUDED.is_active`, m.UserID, m.Username, m.IsActive)
		if err != nil {
			writeErr(w, 500, ErrCodeNotFound, err.Error())
			return
		}
		_, err = tx.Exec(`INSERT INTO team_users(team_name, user_id)
VALUES($1,$2) ON CONFLICT DO NOTHING`, t.TeamName, m.UserID)
		if err != nil {
			writeErr(w, 500, ErrCodeNotFound, err.Error())
			return
		}
	}
	if err := tx.Commit(); err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(teamResp{Team: t})
}

// GET /team/get?team_name=...
func getTeamWithMembers(w http.ResponseWriter, r *http.Request) {
	team := r.URL.Query().Get("team_name")
	if team == "" {
		writeErr(w, 400, ErrCodeNotFound, "team_name required")
		return
	}
	// check team exists
	var exists string
	if err := db.DB.QueryRow(`SELECT name FROM teams WHERE name=$1`,
		team).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			writeErr(w, 404, ErrCodeNotFound, "team not found")
			return
		}
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	rows, err := db.DB.Query(`SELECT u.id, u.name, u.is_active FROM users uJOIN team_users tu ON u.id = tu.user_id WHERE tu.team_name=$1
 `, team)
	if err != nil {
		writeErr(w, 500, ErrCodeNotFound, err.Error())
		return
	}
	defer rows.Close()
	members := []teamMemberReq{}
	for rows.Next() {
		var id, name string
		var isActive bool
		if err := rows.Scan(&id, &name, &isActive); err != nil {
			writeErr(w, 500, ErrCodeNotFound, err.Error())
			return
		}
		members = append(members, teamMemberReq{UserID: id, Username: name,
			IsActive: isActive})
	}
	resp := struct {
		TeamName string          `json:"team_name"`
		Members  []teamMemberReq `json:"members"`
	}{
		TeamName: team,
		Members:  members,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
