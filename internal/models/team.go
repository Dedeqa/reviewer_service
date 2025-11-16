package models

// legacy Team (kept)
type Team struct {
	Name string `json:"name"`
}

// OpenAPI TeamMember
type TeamMember struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

// OpenAPI Team
type TeamOpenAPI struct {
	TeamName string       `json:"team_name"`
	Members  []TeamMember `json:"members"`
}
