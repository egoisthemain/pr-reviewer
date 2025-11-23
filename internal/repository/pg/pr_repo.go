package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/egoisthemain/pr-reviewer/internal/domain"
)

type PRRepository struct {
	db *sql.DB
}

func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}

var (
	ErrPRNotFound = errors.New("pull request not found")
)

func (r *PRRepository) CreatePR(ctx context.Context, pr domain.PullRequest) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
        VALUES ($1, $2, $3, 'OPEN')
    `, pr.PullRequestID, pr.PullRequestName, pr.AuthorID)

	if err != nil {
		return fmt.Errorf("insert pr: %w", err)
	}
	return nil
}

func (r *PRRepository) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, prID)

	var pr domain.PullRequest
	var mergedAt sql.NullTime

	if err := row.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID,
		&pr.Status, &pr.CreatedAt, &mergedAt); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("select pr: %w", err)
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	reviewers, err := r.ListReviewers(ctx, prID)
	if err != nil {
		return nil, fmt.Errorf("select reviewers: %w", err)
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *PRRepository) AddReviewer(ctx context.Context, prID string, userID string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO pr_reviewers (pull_request_id, user_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, prID, userID)

	if err != nil {
		return fmt.Errorf("add reviewer: %w", err)
	}
	return nil
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID string, userID string) error {
	_, err := r.db.ExecContext(ctx, `
        DELETE FROM pr_reviewers
        WHERE pull_request_id = $1 AND user_id = $2
    `, prID, userID)

	if err != nil {
		return fmt.Errorf("remove reviewer: %w", err)
	}
	return nil
}

func (r *PRRepository) ListReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT user_id
        FROM pr_reviewers
        WHERE pull_request_id = $1
    `, prID)
	if err != nil {
		return nil, fmt.Errorf("list reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan reviewer: %w", err)
		}
		reviewers = append(reviewers, uid)
	}
	return reviewers, nil
}

func (r *PRRepository) SetMerged(ctx context.Context, prID string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE pull_requests
        SET status = 'MERGED', merged_at = now()
        WHERE pull_request_id = $1
    `, prID)

	if err != nil {
		return fmt.Errorf("merge pr: %w", err)
	}
	return nil
}

func (r *PRRepository) ListPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
        FROM pr_reviewers r
        JOIN pull_requests pr ON pr.pull_request_id = r.pull_request_id
        WHERE r.user_id = $1
    `, userID)
	if err != nil {
		return nil, fmt.Errorf("list prs by reviewer: %w", err)
	}
	defer rows.Close()

	var list []domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest
		var mergedAt sql.NullTime

		err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID,
			&pr.Status, &pr.CreatedAt, &mergedAt)

		if err != nil {
			return nil, fmt.Errorf("scan pr: %w", err)
		}

		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}

		list = append(list, pr)
	}

	return list, nil
}
