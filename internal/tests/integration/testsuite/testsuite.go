package testsuite

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed fixtures.sql
	fixturesFS embed.FS
)

// Suite wires shared httptest.Server against provided pool and ensures DB reset for each test.
type Suite struct {
	t      *testing.T
	Pool   *pgxpool.Pool
	Server *httptest.Server
}

func New(t *testing.T, pool *pgxpool.Pool, handler http.Handler) *Suite {
	t.Helper()

	loadFixtures(t, pool)

	srv := httptest.NewServer(handler)

	return &Suite{t: t, Pool: pool, Server: srv}
}

func (s *Suite) Close() {
	s.Server.Close()
}

func (s *Suite) doJSON(t *testing.T, method, path string, body any) *http.Response {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, s.Server.URL+path, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := s.Server.Client().Do(req)
	require.NoError(t, err)
	return resp
}

func (s *Suite) PostJSON(t *testing.T, path string, body any) *http.Response {
	return s.doJSON(t, http.MethodPost, path, body)
}

func (s *Suite) Get(t *testing.T, path string) *http.Response {
	return s.doJSON(t, http.MethodGet, path, nil)
}

func (s *Suite) DecodeJSON(t *testing.T, resp *http.Response, dest any) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, dest), "body: %s", string(data))
}

type ErrorPayload struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func loadFixtures(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	data, err := fixturesFS.ReadFile("fixtures.sql")
	require.NoError(t, err)

	_, err = pool.Exec(context.Background(), string(data))
	require.NoError(t, err)
}
