package dto

type ErrorCode string

func (e ErrorCode) Error() string {
	return string(e)
}

const (
	ErrTeamExists     ErrorCode = "TEAM_EXISTS"
	ErrPRExists       ErrorCode = "PR_EXISTS"
	ErrPRMerged       ErrorCode = "PR_MERGED"
	ErrNotAssigned    ErrorCode = "NOT_ASSIGNED"
	ErrNoCandidate    ErrorCode = "NO_CANDIDATE"
	ErrNotFound       ErrorCode = "NOT_FOUND"
	ErrInternalServer ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrBadRequest     ErrorCode = "BAD_REQUEST"
)

type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func NewErrorResponse(code ErrorCode, message string) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorPayload{
			Code:    code,
			Message: message,
		},
	}
}
