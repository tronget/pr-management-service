package team

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) chi.Router {
	repo := NewRepository(pool)
	svc := NewService(repo)
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Post("/add", h.CreateTeam)
	r.Get("/get", h.GetTeam)

	return r
}
