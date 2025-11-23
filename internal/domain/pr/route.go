package pr

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
	r.Post("/create", h.CreatePR)
	r.Post("/merge", h.MergePR)
	r.Post("/reassign", h.ReassignReviewer)

	return r
}
