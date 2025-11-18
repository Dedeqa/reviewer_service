package services

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	"reviewer-service/internal/db"
)

func PickRandomActiveFromTeamExcluding(team string, exclude []string) (string, error) {
	rows, err := db.DB.Query(` SELECT u.id FROM users u JOIN team_users tu ON u.id = tu.user_id WHERE tu.team_name=$1 AND u.is_active=true`, team)
	if err != nil {
		return "", err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}(rows)

	excludeMap := map[string]struct{}{}
	for _, e := range exclude {
		excludeMap[e] = struct{}{}
	}

	var candidates []string

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		if _, excluded := excludeMap[id]; excluded {
			continue
		}
		candidates = append(candidates, id)
	}

	if len(candidates) == 0 {
		return "", nil
	}

	rand.Seed(time.Now().UnixNano())
	return candidates[rand.Intn(len(candidates))], nil
}
