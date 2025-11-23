package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/egoisthemain/pr-reviewer/internal/domain"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

var ErrNotFound = errors.New("not found")
var ErrTeamExists = errors.New("team exists")

func (r *TeamRepository) CreateTeamWithMembers(ctx context.Context, team domain.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`,
		team.TeamName,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check team: %w", err)
	}
	if exists {
		return ErrTeamExists
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO teams (team_name) VALUES ($1)`,
		team.TeamName,
	); err != nil {
		return fmt.Errorf("insert team: %w", err)
	}

	for _, u := range team.Members {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) DO UPDATE
			  SET username = EXCLUDED.username,
			      team_name = EXCLUDED.team_name,
			      is_active = EXCLUDED.is_active
		`, u.UserID, u.Username, team.TeamName, u.IsActive)
		if err != nil {
			return fmt.Errorf("upsert user: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`,
		teamName,
	).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check team: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("select members: %w", err)
	}
	defer rows.Close()

	var members []domain.User
	for rows.Next() {
		var u domain.User
		u.TeamName = teamName
		if err := rows.Scan(&u.UserID, &u.Username, &u.IsActive); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (r *TeamRepository) SetUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`, userID, isActive)

	var u domain.User
	if err := row.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("update: %w", err)
	}
	return &u, nil
}

func (r *TeamRepository) ListTeams(ctx context.Context) ([]domain.Team, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT team_name FROM teams`)
	if err != nil {
		return nil, fmt.Errorf("list teams: %w", err)
	}
	defer rows.Close()

	var out []domain.Team

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan team: %w", err)
		}

		team, err := r.GetTeam(ctx, name)
		if err != nil {
			return nil, err
		}
		out = append(out, *team)
	}

	return out, nil
}

func (r *TeamRepository) AddTeam(ctx context.Context, t domain.Team) error {
	return r.CreateTeamWithMembers(ctx, t)
}
