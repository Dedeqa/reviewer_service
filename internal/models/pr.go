package models

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

// legacy PR (kept)
type PR struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	AuthorID  string   `json:"authorId"`
	TeamName  string   `json:"teamName"`
	Status    PRStatus `json:"status"`
	Reviewers []User   `json:"reviewers"`
}

// OpenAPI PullRequest
type PullRequest struct {
	PullRequestID     string   `json:"pull_request_id"`
	PullRequestName   string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            PRStatus `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
	CreatedAt         *string  `json:"createdAt,omitempty"`
	MergedAt          *string  `json:"mergedAt,omitempty"`
}

// PullRequestShort for /users/getReview response
type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id"`
	PullRequestName string   `json:"pull_request_name"`
	AuthorID        string   `json:"author_id"`
	Status          PRStatus `json:"status"`
}
