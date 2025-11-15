package models

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsActive bool   `json:"isActive"`
}
