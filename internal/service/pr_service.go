package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/egoisthemain/pr-reviewer/internal/domain"
)

type PullRequestRepo interface {
	CreatePR(ctx context.Context, pr domain.PullRequest) error
	GetPR(ctx context.Context, prID string) (*domain.PullRequest, error)
	AddReviewer(ctx context.Context, prID string, userID string) error
	RemoveReviewer(ctx context.Context, prID string, userID string) error
	ListReviewers(ctx context.Context, prID string) ([]string, error)
	SetMerged(ctx context.Context, prID string) error
	ListPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error)
}

type PRService struct {
	prRepo   PullRequestRepo
	teamRepo TeamRepo
}

func NewPRService(prRepo PullRequestRepo, teamRepo TeamRepo) *PRService {
	return &PRService{
		prRepo:   prRepo,
		teamRepo: teamRepo,
	}
}

func (s *PRService) findUser(ctx context.Context, userID string) (*domain.User, string, error) {
	teams, err := s.teamRepo.ListTeams(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("list teams: %w", err)
	}

	for _, t := range teams {
		for _, u := range t.Members {
			if u.UserID == userID {
				return &u, t.TeamName, nil
			}
		}
	}

	return nil, "", fmt.Errorf("user not found")
}

func (s *PRService) pickRandomReviewers(users []domain.User, n int) []domain.User {
	if len(users) <= n {
		return users
	}

	rand.Seed(time.Now().UnixNano())

	out := make([]domain.User, n)
	chosen := make(map[int]bool)

	for i := 0; i < n; {
		idx := rand.Intn(len(users))
		if !chosen[idx] {
			chosen[idx] = true
			out[i] = users[idx]
			i++
		}
	}
	return out
}

func (s *PRService) CreatePR(ctx context.Context, prID, name, authorID string) (*domain.PullRequest, error) {

	_, teamName, err := s.findUser(ctx, authorID)
	if err != nil {
		return nil, fmt.Errorf("author not found: %w", err)
	}

	team, err := s.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return nil, fmt.Errorf("load team: %w", err)
	}

	candidates := make([]domain.User, 0)
	for _, u := range team.Members {
		if u.UserID == authorID {
			continue
		}
		if !u.IsActive {
			continue
		}
		candidates = append(candidates, u)
	}

	selected := s.pickRandomReviewers(candidates, 2)

	pr := domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   name,
		AuthorID:          authorID,
		Status:            domain.PROpen,
		CreatedAt:         time.Now(),
		AssignedReviewers: []string{},
	}

	if err := s.prRepo.CreatePR(ctx, pr); err != nil {
		return nil, fmt.Errorf("create pr: %w", err)
	}

	for _, r := range selected {
		_ = s.prRepo.AddReviewer(ctx, prID, r.UserID)
		pr.AssignedReviewers = append(pr.AssignedReviewers, r.UserID)
	}

	return &pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetPR(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == domain.PRMerged {
		return pr, nil
	}

	if err := s.prRepo.SetMerged(ctx, prID); err != nil {
		return nil, fmt.Errorf("merge pr: %w", err)
	}

	return s.prRepo.GetPR(ctx, prID)
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, error) {
	pr, err := s.prRepo.GetPR(ctx, prID)
	if err != nil {
		return "", err
	}

	if pr.Status == domain.PRMerged {
		return "", fmt.Errorf("PR_MERGED")
	}

	found := false
	for _, r := range pr.AssignedReviewers {
		if r == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("NOT_ASSIGNED")
	}

	_, teamName, err := s.findUser(ctx, oldUserID)
	if err != nil {
		return "", err
	}

	team, err := s.teamRepo.GetTeam(ctx, teamName)
	if err != nil {
		return "", fmt.Errorf("load team: %w", err)
	}

	candidates := make([]domain.User, 0)
	for _, u := range team.Members {
		if !u.IsActive {
			continue
		}
		if u.UserID == oldUserID {
			continue
		}
		candidates = append(candidates, u)
	}

	if len(candidates) == 0 {
		return "", fmt.Errorf("NO_CANDIDATE")
	}

	rand.Seed(time.Now().UnixNano())
	newR := candidates[rand.Intn(len(candidates))]

	_ = s.prRepo.RemoveReviewer(ctx, prID, oldUserID)
	_ = s.prRepo.AddReviewer(ctx, prID, newR.UserID)

	return newR.UserID, nil
}

func (s *PRService) ListPRByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	rows, err := s.prRepo.ListPRsByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
