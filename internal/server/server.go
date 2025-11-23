package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tronget/pr-management-service/internal/config"
	"github.com/tronget/pr-management-service/internal/domain/pr"
	"github.com/tronget/pr-management-service/internal/domain/team"
	"github.com/tronget/pr-management-service/internal/domain/user"
)

type Server interface {
	Start() error
}

type server struct {
	cfg    *config.Config
	dbPool *pgxpool.Pool
}

func New(cfg *config.Config, dbPool *pgxpool.Pool) Server {
	return &server{
		cfg:    cfg,
		dbPool: dbPool,
	}
}

func (s *server) Start() error {
	r := Handler(s.dbPool)
	return http.ListenAndServe(s.cfg.HTTP.Address, r)
}

func Handler(pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/team", team.NewRouter(pool))
	r.Mount("/users", user.NewRouter(pool))
	r.Mount("/pullRequest", pr.NewRouter(pool))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return r
}
