package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AlertGroupHandler struct {
	svc *service.AlertGroupService
}

func NewAlertGroupHandler(svc *service.AlertGroupService) *AlertGroupHandler {
	return &AlertGroupHandler{svc: svc}
}

func (h *AlertGroupHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/members", h.AddMember)
	r.Get("/{id}/members", h.ListMembers)
	r.Delete("/{id}/members/{userId}", h.RemoveMember)
	return r
}

func (h *AlertGroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name              string  `json:"name"`
		Description       string  `json:"description"`
		ParentID          *string `json:"parent_id"`
		EscalationMinutes *int    `json:"escalation_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	g, err := h.svc.Create(r.Context(), req.Name, req.Description, req.ParentID, req.EscalationMinutes)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, g)
}

func (h *AlertGroupHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	g, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, g)
}

func (h *AlertGroupHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, items)
}

func (h *AlertGroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name              string  `json:"name"`
		Description       string  `json:"description"`
		ParentID          *string `json:"parent_id"`
		EscalationMinutes *int    `json:"escalation_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	g, err := h.svc.Update(r.Context(), id, req.Name, req.Description, req.ParentID, req.EscalationMinutes)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, g)
}

func (h *AlertGroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *AlertGroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.AddMember(r.Context(), groupID, req.UserID, req.Role)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, m)
}

func (h *AlertGroupHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	members, err := h.svc.ListMembers(r.Context(), groupID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, members)
}

func (h *AlertGroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	if err := h.svc.RemoveMember(r.Context(), groupID, userID); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}
