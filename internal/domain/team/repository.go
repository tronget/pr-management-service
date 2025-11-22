package team

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
	CreateTeam(ctx context.Context, teamName string, members []Member) (*Team, error)
	GetTeam(ctx context.Context, teamName string) (*Team, error)
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{db: pool}
}

type repository struct {
	db *pgxpool.Pool
}

func (r *repository) CreateTeam(ctx context.Context, teamName string, members []Member) (*Team, error) {
	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	defer db.FinalizeTx(ctx, tx, &err)

	if _, err = tx.Exec(ctx, `INSERT INTO teams (team_name) VALUES ($1)`, teamName); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == db.PgUniqueViolation {
			return nil, dto.ErrTeamExists
		}
		return nil, err
	}

	const upsertMember = `INSERT INTO users (user_id, username, team_name, is_active)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE
	SET username = EXCLUDED.username,
		team_name = EXCLUDED.team_name,
		is_active = EXCLUDED.is_active
RETURNING user_id, username, is_active`

	if len(members) == 0 {
		team := &Team{Name: teamName, Members: []Member{}}
		return team, nil
	}

	batch := &pgx.Batch{}
	for _, member := range members {
		batch.Queue(upsertMember, member.UserID, member.Username, teamName, member.IsActive)
	}

	br := tx.SendBatch(ctx, batch)
	defer func() {
		err2 := br.Close()
		if err == nil {
			err = err2
		}
	}()

	savedMembers := make([]Member, 0, len(members))
	for i := 0; i < len(members); i++ {
		var m Member
		if err = br.QueryRow().Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		savedMembers = append(savedMembers, m)
	}

	team := &Team{Name: teamName, Members: savedMembers}
	return team, nil
}

func (r *repository) GetTeam(ctx context.Context, teamName string) (*Team, error) {
	var exists bool
	if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM teams WHERE team_name = $1)`, teamName).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, dto.ErrNotFound
	}

	rows, err := r.db.Query(ctx, `SELECT user_id, username, is_active FROM users WHERE team_name = $1 ORDER BY user_id`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]Member, 0)
	for rows.Next() {
		var m Member
		if err = rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &Team{Name: teamName, Members: members}, nil
}
