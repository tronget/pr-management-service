package pr

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

func (h *Handler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "invalid payload"),
		)
		return
	}
	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "pull_request_id, pull_request_name and author_id are required"),
		)
		return
	}

	resp, err := h.service.CreatePR(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			httpx.WriteJSON(
				w,
				http.StatusNotFound,
				dto.NewErrorResponse(dto.ErrNotFound, "resource not found"),
			)
		case errors.Is(err, dto.ErrPRExists):
			httpx.WriteJSON(
				w,
				http.StatusConflict,
				dto.NewErrorResponse(dto.ErrPRExists, "pull request already exists"),
			)
		default:
			log.Printf("CreatePR: unexpected error: %v", err)
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

func (h *Handler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "invalid payload"),
		)
		return
	}
	if req.PullRequestID == "" {
		httpx.WriteJSON(
			w,
			http.StatusBadRequest,
			dto.NewErrorResponse(dto.ErrBadRequest, "pull_request_id is required"),
		)
		return
	}

	resp, err := h.service.MergePR(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			httpx.WriteJSON(
				w,
				http.StatusNotFound,
				dto.NewErrorResponse(dto.ErrNotFound, "resource not found"),
			)
		default:
			log.Printf("MergePR: unexpected error: %v", err)
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

func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	var req ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, dto.NewErrorResponse(dto.ErrBadRequest, "invalid payload"))
		return
	}
	if req.PullRequestID == "" || req.OldUserID == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, dto.NewErrorResponse(dto.ErrBadRequest, "pull_request_id and old_user_id are required"))
		return
	}

	resp, err := h.service.ReassignReviewer(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, dto.ErrNotFound):
			httpx.WriteJSON(
				w,
				http.StatusNotFound,
				dto.NewErrorResponse(dto.ErrNotFound, "resource not found"),
			)
		case errors.Is(err, dto.ErrPRMerged):
			httpx.WriteJSON(
				w,
				http.StatusConflict,
				dto.NewErrorResponse(dto.ErrPRMerged, "cannot reassign on merged PR"),
			)
		case errors.Is(err, dto.ErrNotAssigned):
			httpx.WriteJSON(
				w,
				http.StatusConflict,
				dto.NewErrorResponse(dto.ErrNotAssigned, "reviewer is not assigned to this PR"),
			)
		case errors.Is(err, dto.ErrNoCandidate):
			httpx.WriteJSON(
				w,
				http.StatusConflict,
				dto.NewErrorResponse(dto.ErrNoCandidate, "no active replacement candidate in team"),
			)
		default:
			log.Printf("ReassignReviewer: unexpected error: %v", err)
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
