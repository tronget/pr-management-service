package pr

import "time"

type Status string

const (
	StatusOpen   Status = "OPEN"
	StatusMerged Status = "MERGED"
)

type PullRequest struct {
	ID                string     `db:"pull_request_id" json:"pull_request_id"`
	Name              string     `db:"pull_request_name" json:"pull_request_name"`
	AuthorID          string     `db:"author_id" json:"author_id"`
	Status            Status     `db:"status" json:"status"`
	AssignedReviewers []string   `db:"-" json:"assigned_reviewers"`
	CreatedAt         time.Time  `db:"created_at" json:"createdAt,omitempty"`
	MergedAt          *time.Time `db:"merged_at" json:"mergedAt,omitempty"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}
