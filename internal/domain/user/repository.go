package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tronget/pr-management-service/internal/common/dto"
	"github.com/tronget/pr-management-service/internal/domain/pr"
)

type Repository interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*User, error)
	GetReviewerPRs(ctx context.Context, userID string) ([]pr.PullRequestShort, error)
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{db: pool}
}

type repository struct {
	db *pgxpool.Pool
}

func (r *repository) SetIsActive(ctx context.Context, userID string, isActive bool) (*User, error) {
	const query = `UPDATE users SET is_active = $2 WHERE user_id = $1
RETURNING user_id, username, team_name, is_active`

	var u User
	if err := r.db.QueryRow(ctx, query, userID, isActive).
		Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *repository) GetReviewerPRs(ctx context.Context, userID string) ([]pr.PullRequestShort, error) {
	if err := r.ensureUserExists(ctx, userID); err != nil {
		return nil, err
	}

	const query = `SELECT p.pull_request_id, p.pull_request_name, p.author_id, p.status
FROM pull_requests p
JOIN pull_request_reviewers prr ON prr.pull_request_id = p.pull_request_id
WHERE prr.reviewer_id = $1
ORDER BY p.created_at DESC, p.pull_request_id`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prs := make([]pr.PullRequestShort, 0)
	for rows.Next() {
		var item pr.PullRequestShort
		if err = rows.Scan(&item.PullRequestID, &item.PullRequestName, &item.AuthorID, &item.Status); err != nil {
			return nil, err
		}
		prs = append(prs, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

func (r *repository) ensureUserExists(ctx context.Context, userID string) error {
	const query = `SELECT user_id FROM users WHERE user_id = $1`

	var id string
	if err := r.db.QueryRow(ctx, query, userID).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.ErrNotFound
		}
		return err
	}
	return nil
}
