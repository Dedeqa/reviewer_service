package services

import (
	"math/rand"

	"reviewer-service/internal/db"
	"reviewer-service/internal/models"
)

func LoadPR(id string) (models.PR, error) {
	var p models.PR
	var status string

	row := db.DB.QueryRow(`
        SELECT id,title,author_id,team_name,status
        FROM prs WHERE id=$1
    `, id)

	if err := row.Scan(&p.ID, &p.Title, &p.AuthorID, &p.TeamName, &status); err != nil {
		return p, err
	}

	p.Status = models.PRStatus(status)

	rows, err := db.DB.Query(`
        SELECT u.id,u.name,u.is_active
        FROM users u
        JOIN pr_reviewers prr ON u.id=prr.reviewer_id
        WHERE prr.pr_id=$1
    `, id)
	if err != nil {
		return p, err
	}
	defer rows.Close()

	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.IsActive); err != nil {
			return p, err
		}
		p.Reviewers = append(p.Reviewers, u)
	}

	return p, nil
}

func AssignUpToTwoReviewers(prID, authorID, team string) error {
	rows, err := db.DB.Query(`
        SELECT u.id
        FROM users u
        JOIN team_users tu ON u.id = tu.user_id
        WHERE tu.team_name=$1 AND u.is_active=true AND u.id<>$2
    `, team, authorID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var candidates []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		candidates = append(candidates, id)
	}

	if len(candidates) == 0 {
		return nil
	}

	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	count := 2
	if len(candidates) < 2 {
		count = len(candidates)
	}

	for i := 0; i < count; i++ {
		_, err := db.DB.Exec(`
            INSERT INTO pr_reviewers(pr_id, reviewer_id) 
            VALUES($1,$2)
        `, prID, candidates[i])
		if err != nil {
			return err
		}
	}

	return nil
}
