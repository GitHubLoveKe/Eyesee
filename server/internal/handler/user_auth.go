package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UserAuthHandler struct {
	svc *service.UserAuthService
}

func NewUserAuthHandler(svc *service.UserAuthService) *UserAuthHandler {
	return &UserAuthHandler{svc: svc}
}

func (h *UserAuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/password", h.ChangePassword)
	r.Get("/{userId}/notification-targets", h.ListNotificationTargets)
	r.Post("/{userId}/notification-targets", h.AddNotificationTarget)
	r.Delete("/notification-targets/{id}", h.RemoveNotificationTarget)
	r.Post("/{userId}/notification-targets/{id}/primary", h.SetPrimaryTarget)
	return r
}

func (h *UserAuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string `json:"user_id"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if err := h.svc.ChangePassword(r.Context(), req.UserID, req.NewPassword); err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "updated"})
}

func (h *UserAuthHandler) ListNotificationTargets(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	targets, err := h.svc.ListNotificationTargets(r.Context(), userID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, targets)
}

func (h *UserAuthHandler) AddNotificationTarget(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	var req struct {
		ChannelType string `json:"channel_type"`
		TargetValue string `json:"target_value"`
		IsPrimary   bool   `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	t, err := h.svc.AddNotificationTarget(r.Context(), service.AddNotificationTargetRequest{
		UserID:      userID,
		ChannelType: req.ChannelType,
		TargetValue: req.TargetValue,
		IsPrimary:   req.IsPrimary,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, t)
}

func (h *UserAuthHandler) RemoveNotificationTarget(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.RemoveNotificationTarget(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *UserAuthHandler) SetPrimaryTarget(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid id")
		return
	}
	if err := h.svc.SetPrimaryTarget(r.Context(), userID, id); err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "updated"})
}
