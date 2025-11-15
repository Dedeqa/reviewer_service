package models

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PR struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	AuthorID  string   `json:"authorId"`
	TeamName  string   `json:"teamName"`
	Status    PRStatus `json:"status"`
	Reviewers []User   `json:"reviewers"`
}
