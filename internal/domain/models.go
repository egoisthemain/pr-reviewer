package domain

import "time"

type User struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name,omitempty"`
	IsActive bool   `json:"is_active"`
}

type Team struct {
	TeamName string `json:"team_name"`
	Members  []User `json:"members"`
}

type PRStatus string

const (
	PROpen   PRStatus = "OPEN"
	PRMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	MergedAt          *time.Time `json:"merged_at,omitempty"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
}
