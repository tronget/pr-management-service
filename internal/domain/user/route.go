package user

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool) http.Handler {
	repo := NewRepository(pool)
	svc := NewService(repo)
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Post("/setIsActive", h.SetIsActive)
	r.Get("/getReview", h.GetReview)

	return r
}
