package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tronget/pr-management-service/internal/server"
	"github.com/tronget/pr-management-service/internal/tests/integration/testsuite"
)

type prResponse struct {
	PR struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_id"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
		MergedAt          *string  `json:"mergedAt"`
	} `json:"pr"`
}

type reassignResponse struct {
	PR struct {
		PullRequestID     string   `json:"pull_request_id"`
		PullRequestName   string   `json:"pull_request_name"`
		AuthorID          string   `json:"author_id"`
		Status            string   `json:"status"`
		AssignedReviewers []string `json:"assigned_reviewers"`
	} `json:"pr"`
	ReplacedBy string `json:"replaced_by"`
}

func TestPullRequestFlow(t *testing.T) {
	pool := dbPool(t)
	suite := testsuite.New(t, pool, server.Handler(pool))
	t.Cleanup(suite.Close)

	// Create PR
	createPayload := map[string]any{
		"pull_request_id":   "pr-1",
		"pull_request_name": "Add feature",
		"author_id":         "alpha-author",
	}
	resp := suite.PostJSON(t, "/pullRequest/create", createPayload)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created prResponse
	suite.DecodeJSON(t, resp, &created)
	require.Equal(t, "pr-1", created.PR.PullRequestID)
	require.Equal(t, "OPEN", created.PR.Status)
	// Should assign up to 2 reviewers, none should be author or inactive.
	require.LessOrEqual(t, len(created.PR.AssignedReviewers), 2)
	for _, rID := range created.PR.AssignedReviewers {
		require.NotEqual(t, "alpha-author", rID)
		require.NotEqual(t, "alpha-inactive", rID)
	}

	// Merge PR (first time)
	mergePayload := map[string]any{"pull_request_id": "pr-1"}
	mergeResp := suite.PostJSON(t, "/pullRequest/merge", mergePayload)
	require.Equal(t, http.StatusOK, mergeResp.StatusCode)
	var merged1 prResponse
	suite.DecodeJSON(t, mergeResp, &merged1)
	require.Equal(t, "MERGED", merged1.PR.Status)
	require.NotNil(t, merged1.PR.MergedAt)

	// Merge PR again (idempotent)
	mergeResp2 := suite.PostJSON(t, "/pullRequest/merge", mergePayload)
	require.Equal(t, http.StatusOK, mergeResp2.StatusCode)
	var merged2 prResponse
	suite.DecodeJSON(t, mergeResp2, &merged2)
	require.Equal(t, "MERGED", merged2.PR.Status)
	require.Equal(t, merged1.PR.PullRequestID, merged2.PR.PullRequestID)

	// Reassign reviewer on merged PR should fail with PR_MERGED
	reassignPayloadMerged := map[string]any{
		"pull_request_id": "pr-1",
		"old_user_id": func() string {
			if len(merged2.PR.AssignedReviewers) > 0 {
				return merged2.PR.AssignedReviewers[0]
			}
			return "alpha-reviewer-1"
		}(),
	}
	reassMergedResp := suite.PostJSON(t, "/pullRequest/reassign", reassignPayloadMerged)
	require.Equal(t, http.StatusConflict, reassMergedResp.StatusCode)
	var errPayload testsuite.ErrorPayload
	suite.DecodeJSON(t, reassMergedResp, &errPayload)
	require.Equal(t, "PR_MERGED", errPayload.Error.Code)

	// Create second PR to test reassignment success
	createPayload2 := map[string]any{
		"pull_request_id":   "pr-2",
		"pull_request_name": "Another feature",
		"author_id":         "alpha-author",
	}
	resp2 := suite.PostJSON(t, "/pullRequest/create", createPayload2)
	require.Equal(t, http.StatusCreated, resp2.StatusCode)
	var created2 prResponse
	suite.DecodeJSON(t, resp2, &created2)
	require.NotEmpty(t, created2.PR.AssignedReviewers)

	// Pick first reviewer to replace
	oldReviewer := created2.PR.AssignedReviewers[0]
	reassignPayload := map[string]any{
		"pull_request_id": "pr-2",
		"old_user_id":     oldReviewer,
	}
	reassResp := suite.PostJSON(t, "/pullRequest/reassign", reassignPayload)
	require.Equal(t, http.StatusOK, reassResp.StatusCode)
	var reassigned reassignResponse
	suite.DecodeJSON(t, reassResp, &reassigned)
	require.Equal(t, "pr-2", reassigned.PR.PullRequestID)
	require.NotEqual(t, oldReviewer, reassigned.ReplacedBy)
	// Ensure old reviewer removed
	for _, rID := range reassigned.PR.AssignedReviewers {
		require.NotEqual(t, oldReviewer, rID)
	}
}
