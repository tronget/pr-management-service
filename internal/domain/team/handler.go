package team

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/tronget/pr-management-service/internal/common/dto"
	"github.com/tronget/pr-management-service/pkg/httpx"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "invalid payload"),
		)
		return
	}
	if req.TeamName == "" {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "team_name is required"),
		)
		return
	}

	resp, err := h.service.CreateTeam(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrTeamExists):
			httpx.WriteJSON(
				w,
				http.StatusBadRequest,
				dto.NewErrorResponse(dto.ErrTeamExists, "team_name already exists"),
			)
		default:
			log.Printf("GetTeam: unexpected error: %v", err)
			httpx.WriteJSON(
				w,
				http.StatusInternalServerError,
				dto.NewErrorResponse(dto.ErrInternalServer, "internal server error"),
			)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "team_name is required"),
		)
		return
	}

	resp, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			httpx.WriteJSON(
				w,
				http.StatusNotFound,
				dto.NewErrorResponse(dto.ErrNotFound, "resource not found"),
			)
		default:
			log.Printf("GetTeam: unexpected error: %v", err)
			httpx.WriteJSON(
				w,
				http.StatusInternalServerError,
				dto.NewErrorResponse(dto.ErrInternalServer, "internal server error"),
			)
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, resp)
}
