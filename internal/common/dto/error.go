package dto

type ErrorCode string

const (
	ErrTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrPRExists    ErrorCode = "PR_EXISTS"
	ErrPRMerged    ErrorCode = "PR_MERGED"
	ErrNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrNotFound    ErrorCode = "NOT_FOUND"
)

type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}
