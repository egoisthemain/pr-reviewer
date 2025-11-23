package service

import (
	"context"

	"github.com/egoisthemain/pr-reviewer/internal/domain"
)

type TeamRepo interface {
	AddTeam(ctx context.Context, team domain.Team) error
	CreateTeamWithMembers(ctx context.Context, team domain.Team) error
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
	ListTeams(ctx context.Context) ([]domain.Team, error)
	SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
}

type TeamService struct {
	repo TeamRepo
}

func NewTeamService(repo TeamRepo) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) CreateTeam(ctx context.Context, t domain.Team) error {
	return s.repo.CreateTeamWithMembers(ctx, t)
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	return s.repo.GetTeam(ctx, teamName)
}

func (s *TeamService) ListTeams(ctx context.Context) ([]domain.Team, error) {
	return s.repo.ListTeams(ctx)
}

func (s *TeamService) SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	return s.repo.SetUserActive(ctx, userID, isActive)
}
