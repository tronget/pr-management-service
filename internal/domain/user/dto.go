package user

import "github.com/tronget/pr-management-service/internal/domain/pr"

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	User User `json:"user"`
}

type GetReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []pr.PullRequestShort `json:"pull_requests"`
}
