package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tronget/pr-management-service/internal/config"
)

type DB interface {
	Connect(ctx context.Context) (*pgxpool.Pool, error)
	Close()
}

type db struct {
	dsn  string
	pool *pgxpool.Pool
	mu   sync.Mutex
}

func (d *db) Connect(ctx context.Context) (*pgxpool.Pool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.pool != nil {
		return d.pool, nil
	}

	cfg, err := pgxpool.ParseConfig(d.dsn)
	if err != nil {
		return nil, fmt.Errorf("pgxpool parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("initializing new pgxpool with config: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}

	d.pool = pool
	return d.pool, nil
}

func (d *db) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.pool != nil {
		d.pool.Close()
		d.pool = nil
	}
}

func New(cfg *config.Config) DB {
	return &db{dsn: cfg.DbUrl}
}
