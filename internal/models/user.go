package models

// legacy User (kept)
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
	// TeamName included as optional field when scanning
	TeamName string `json:"teamName,omitempty"`
}

// OpenAPI User
type UserOpenAPI struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}
