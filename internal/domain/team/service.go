package team

import "context"

type Service interface {
	CreateTeam(ctx context.Context, req CreateTeamRequest) (*CreateTeamResponse, error)
	GetTeam(ctx context.Context, teamName string) (*GetTeamResponse, error)
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

type service struct {
	repo Repository
}

func (s *service) CreateTeam(ctx context.Context, req CreateTeamRequest) (*CreateTeamResponse, error) {
	members := append([]Member(nil), req.Members...)

	team, err := s.repo.CreateTeam(ctx, req.TeamName, members)
	if err != nil {
		return nil, err
	}

	return &CreateTeamResponse{Team: *team}, nil
}

func (s *service) GetTeam(ctx context.Context, teamName string) (*GetTeamResponse, error) {
	team, err := s.repo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	return &GetTeamResponse{TeamName: team.Name, Members: team.Members}, nil
}
