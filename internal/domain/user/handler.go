package user

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

func (h *Handler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "invalid payload"),
		)
		return
	}
	if req.UserID == "" {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "user_id is required"),
		)
		return
	}

	resp, err := h.service.SetIsActive(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			httpx.WriteJSON(
				w,
				http.StatusNotFound,
				dto.NewErrorResponse(dto.ErrNotFound, "resource not found"),
			)
		default:
			log.Printf("SetIsActive: unexpected error: %v", err)
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

func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "user_id is required"),
		)
		return
	}

	resp, err := h.service.GetReview(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			httpx.WriteJSON(
				w,
				http.StatusNotFound,
				dto.NewErrorResponse(dto.ErrNotFound, "resource not found"),
			)
		default:
			log.Printf("GetReview: unexpected error: %v", err)
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
