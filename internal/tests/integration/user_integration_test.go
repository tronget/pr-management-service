package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tronget/pr-management-service/internal/server"
	"github.com/tronget/pr-management-service/internal/tests/integration/testsuite"
)

type setActiveResp struct {
	User struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		TeamName string `json:"team_name"`
		IsActive bool   `json:"is_active"`
	} `json:"user"`
}

type reviewResp struct {
	UserID       string `json:"user_id"`
	PullRequests []struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
		Status          string `json:"status"`
	} `json:"pull_requests"`
}

func TestUserEndpoints(t *testing.T) {
	pool := dbPool(t)
	suite := testsuite.New(t, pool, server.Handler(pool))
	t.Cleanup(suite.Close)

	// Toggle user inactive -> active
	payload := map[string]any{"user_id": "alpha-inactive", "is_active": true}
	resp := suite.PostJSON(t, "/users/setIsActive", payload)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var setResp setActiveResp
	suite.DecodeJSON(t, resp, &setResp)
	require.True(t, setResp.User.IsActive)

	// Create a PR so reviewer listing will have entries
	createPR := map[string]any{"pull_request_id": "pr-user-1", "pull_request_name": "User flow", "author_id": "alpha-author"}
	prResp := suite.PostJSON(t, "/pullRequest/create", createPR)
	require.Equal(t, http.StatusCreated, prResp.StatusCode)

	// Get review list for first assigned reviewer
	var created struct {
		PR struct {
			AssignedReviewers []string `json:"assigned_reviewers"`
		} `json:"pr"`
	}
	suite.DecodeJSON(t, prResp, &created)
	if len(created.PR.AssignedReviewers) == 0 {
		t.Skip("no reviewers assigned due to limited active users")
	}
	reviewer := created.PR.AssignedReviewers[0]

	listResp := suite.Get(t, "/users/getReview?user_id="+reviewer)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var review reviewResp
	suite.DecodeJSON(t, listResp, &review)
	found := false
	for _, item := range review.PullRequests {
		if item.PullRequestID == "pr-user-1" {
			found = true
			break
		}
	}
	require.True(t, found, "expected PR in reviewer list")
}
