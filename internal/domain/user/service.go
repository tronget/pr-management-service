package user

import "context"

type Service interface {
	SetIsActive(ctx context.Context, req SetIsActiveRequest) (*SetIsActiveResponse, error)
	GetReview(ctx context.Context, userID string) (*GetReviewResponse, error)
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

type service struct {
	repo Repository
}

func (s *service) SetIsActive(ctx context.Context, req SetIsActiveRequest) (*SetIsActiveResponse, error) {
	user, err := s.repo.SetIsActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		return nil, err
	}

	return &SetIsActiveResponse{User: *user}, nil
}

func (s *service) GetReview(ctx context.Context, userID string) (*GetReviewResponse, error) {
	prs, err := s.repo.GetReviewerPRs(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &GetReviewResponse{UserID: userID, PullRequests: prs}, nil
}
