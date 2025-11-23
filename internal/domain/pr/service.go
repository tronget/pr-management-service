package pr

import "context"

type Service interface {
	CreatePR(ctx context.Context, req CreatePRRequest) (*CreatePRResponse, error)
	MergePR(ctx context.Context, req MergePRRequest) (*MergePRResponse, error)
	ReassignReviewer(ctx context.Context, req ReassignRequest) (*ReassignResponse, error)
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

type service struct {
	repo Repository
}

func (s *service) CreatePR(ctx context.Context, req CreatePRRequest) (*CreatePRResponse, error) {
	pr, err := s.repo.CreatePR(ctx, req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		return nil, err
	}

	return &CreatePRResponse{PR: *pr}, nil
}

func (s *service) MergePR(ctx context.Context, req MergePRRequest) (*MergePRResponse, error) {
	pr, err := s.repo.MergePR(ctx, req.PullRequestID)
	if err != nil {
		return nil, err
	}

	return &MergePRResponse{PR: *pr}, nil
}

func (s *service) ReassignReviewer(ctx context.Context, req ReassignRequest) (*ReassignResponse, error) {
	pr, replacedBy, err := s.repo.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		return nil, err
	}

	return &ReassignResponse{PR: *pr, ReplacedBy: replacedBy}, nil
}
