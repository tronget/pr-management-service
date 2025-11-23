package integration

import (
	"context"
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	testDBURL string
	once      sync.Once
	pool      *pgxpool.Pool
)

func init() {
	flag.StringVar(&testDBURL, "integration-db", os.Getenv("INTEGRATION_DB_URL"), "integration DB URL")
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testDBURL == "" {
		testDBURL = os.Getenv("DB_URL")
	}
	if testDBURL == "" {
		panic("set INTEGRATION_DB_URL or DB_URL for integration tests")
	}

	code := m.Run()
	if pool != nil {
		pool.Close()
	}
	os.Exit(code)
}

func dbPool(t *testing.T) *pgxpool.Pool {
	once.Do(func() {
		var err error
		pool, err = pgxpool.New(context.Background(), testDBURL)
		require.NoError(t, err)
		require.NoError(t, pool.Ping(context.Background()))
	})
	return pool
}
