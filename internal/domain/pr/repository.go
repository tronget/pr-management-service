package pr

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tronget/pr-management-service/internal/common/dto"
	"github.com/tronget/pr-management-service/pkg/db"
)

type Repository interface {
	CreatePR(ctx context.Context, pullRequestID, name, authorID string) (*PullRequest, error)
	MergePR(ctx context.Context, pullRequestID string) (*PullRequest, error)
	ReassignReviewer(ctx context.Context, pullRequestID, oldReviewerID string) (*PullRequest, string, error)
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{db: pool}
}

type repository struct {
	db *pgxpool.Pool
}

func (r *repository) CreatePR(ctx context.Context, pullRequestID, name, authorID string) (pr *PullRequest, err error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer db.FinalizeTx(ctx, tx, &err)

	var teamName string
	if err = tx.QueryRow(ctx, `SELECT team_name FROM users WHERE user_id = $1`, authorID).Scan(&teamName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	const insertPR = `INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id)
VALUES ($1, $2, $3)
RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`

	var created PullRequest
	if err = tx.QueryRow(ctx, insertPR, pullRequestID, name, authorID).
		Scan(&created.ID, &created.Name, &created.AuthorID, &created.Status, &created.CreatedAt, &created.MergedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.PgUniqueViolation {
			return nil, dto.ErrPRExists
		}
		return nil, err
	}

	reviewers, err := r.selectActiveReviewers(ctx, tx, teamName, authorID, 2)
	if err != nil {
		return nil, err
	}

	if len(reviewers) > 0 {
		batch := &pgx.Batch{}
		for _, reviewerID := range reviewers {
			batch.Queue(
				`INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`,
				pullRequestID,
				reviewerID,
			)
		}
		br := tx.SendBatch(ctx, batch)
		if err = br.Close(); err != nil {
			return nil, err
		}
	}

	created.AssignedReviewers = reviewers
	return &created, nil
}

func (r *repository) MergePR(ctx context.Context, pullRequestID string) (pr *PullRequest, err error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer db.FinalizeTx(ctx, tx, &err)

	current, err := r.fetchPR(ctx, tx, pullRequestID)
	if err != nil {
		return nil, err
	}

	if current.Status == StatusMerged {
		return current, nil
	}

	const mergeQuery = `UPDATE pull_requests SET status = 'MERGED', merged_at = NOW()
WHERE pull_request_id = $1
RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`

	var merged PullRequest
	if err = tx.QueryRow(ctx, mergeQuery, pullRequestID).
		Scan(&merged.ID, &merged.Name, &merged.AuthorID, &merged.Status, &merged.CreatedAt, &merged.MergedAt); err != nil {
		return nil, err
	}

	reviewers, err := r.listReviewers(ctx, tx, pullRequestID)
	if err != nil {
		return nil, err
	}
	merged.AssignedReviewers = reviewers

	return &merged, nil
}

func (r *repository) ReassignReviewer(ctx context.Context, pullRequestID, oldReviewerID string) (pr *PullRequest, newReviewer string, err error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, "", err
	}
	defer db.FinalizeTx(ctx, tx, &err)

	current, err := r.fetchPR(ctx, tx, pullRequestID)
	if err != nil {
		return nil, "", err
	}
	if current.Status == StatusMerged {
		return nil, "", dto.ErrPRMerged
	}

	var reviewerTeam string
	const reviewerTeamQuery = `SELECT u.team_name
FROM pull_request_reviewers prr
JOIN users u ON u.user_id = prr.reviewer_id
WHERE prr.pull_request_id = $1 AND prr.reviewer_id = $2`
	if err = tx.QueryRow(ctx, reviewerTeamQuery, pullRequestID, oldReviewerID).Scan(&reviewerTeam); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", dto.ErrNotAssigned
		}
		return nil, "", err
	}

	var replacement string
	replacement, err = r.pickReplacement(ctx, tx, reviewerTeam, current.AuthorID, pullRequestID, oldReviewerID)
	if err != nil {
		return nil, "", err
	}

	if _, err = tx.Exec(ctx, `DELETE FROM pull_request_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2`, pullRequestID, oldReviewerID); err != nil {
		return nil, "", err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`, pullRequestID, replacement); err != nil {
		return nil, "", err
	}

	updated, err := r.fetchPR(ctx, tx, pullRequestID)
	if err != nil {
		return nil, "", err
	}

	return updated, replacement, nil
}

func (r *repository) fetchPR(ctx context.Context, tx pgx.Tx, pullRequestID string) (*PullRequest, error) {
	const selectPR = `SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
FROM pull_requests WHERE pull_request_id = $1`

	var pr PullRequest
	if err := tx.QueryRow(ctx, selectPR, pullRequestID).
		Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dto.ErrNotFound
		}
		return nil, err
	}

	reviewers, err := r.listReviewers(ctx, tx, pullRequestID)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers
	return &pr, nil
}

func (r *repository) listReviewers(ctx context.Context, tx pgx.Tx, pullRequestID string) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT reviewer_id FROM pull_request_reviewers WHERE pull_request_id = $1 ORDER BY assigned_at, reviewer_id`, pullRequestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var reviewer string
		if err = rows.Scan(&reviewer); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (r *repository) selectActiveReviewers(ctx context.Context, tx pgx.Tx, teamName, authorID string, limit int) ([]string, error) {
	rows, err := tx.Query(ctx, `SELECT user_id FROM users WHERE team_name = $1 AND user_id <> $2 AND is_active = TRUE ORDER BY random() LIMIT $3`, teamName, authorID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviewers := make([]string, 0, limit)
	for rows.Next() {
		var reviewer string
		if err = rows.Scan(&reviewer); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewer)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (r *repository) pickReplacement(ctx context.Context, tx pgx.Tx, teamName, authorID, pullRequestID, exclude string) (string, error) {
	const candidateQuery = `SELECT user_id FROM users WHERE team_name = $1 AND user_id <> $2 AND user_id <> $3 AND is_active = TRUE
AND NOT EXISTS (
    SELECT 1 FROM pull_request_reviewers prr WHERE prr.pull_request_id = $4 AND prr.reviewer_id = users.user_id
)
ORDER BY random() LIMIT 1`

	var candidate string
	if err := tx.QueryRow(ctx, candidateQuery, teamName, exclude, authorID, pullRequestID).Scan(&candidate); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", dto.ErrNoCandidate
		}
		return "", err
	}
	return candidate, nil
}
