package server

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tronget/pr-management-service/internal/config"
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

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return http.ListenAndServe(s.cfg.HTTP.Address, r)
}
