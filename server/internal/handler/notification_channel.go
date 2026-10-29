package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type NotificationChannelHandler struct {
	svc *service.NotificationChannelService
}

func NewNotificationChannelHandler(svc *service.NotificationChannelService) *NotificationChannelHandler {
	return &NotificationChannelHandler{svc: svc}
}

func (h *NotificationChannelHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/test", h.Test)
	return r
}

func (h *NotificationChannelHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		ChannelType string `json:"channel_type"`
		Config      string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	nc, err := h.svc.Create(r.Context(), req.Name, req.ChannelType, req.Config)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, nc)
}

func (h *NotificationChannelHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	nc, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, nc)
}

func (h *NotificationChannelHandler) List(w http.ResponseWriter, r *http.Request) {
	channelType := r.URL.Query().Get("channel_type")
	items, err := h.svc.List(r.Context(), channelType)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, items)
}

func (h *NotificationChannelHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name        string `json:"name"`
		ChannelType string `json:"channel_type"`
		Config      string `json:"config"`
		Enabled     *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	nc, err := h.svc.Update(r.Context(), id, req.Name, req.ChannelType, req.Config, req.Enabled)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, nc)
}

func (h *NotificationChannelHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *NotificationChannelHandler) Test(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Test(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "tested"})
}
