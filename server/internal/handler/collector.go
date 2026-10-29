package handler

import (
	"dcs/internal/model"
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type CollectorHandler struct {
	svc *service.CollectorService
}

func NewCollectorHandler(svc *service.CollectorService) *CollectorHandler {
	return &CollectorHandler{svc: svc}
}

func (h *CollectorHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Register)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/heartbeat", h.Heartbeat)

	return r
}

func (h *CollectorHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID         string `json:"zone_id"`
		Name           string `json:"name"`
		NodeType       string `json:"node_type"`
		Address        string `json:"address"`
		HealthEndpoint string `json:"health_endpoint"`
		Labels         string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	n := &model.CollectorNode{
		ZoneID:         req.ZoneID,
		Name:           req.Name,
		NodeType:       req.NodeType,
		Address:        req.Address,
		HealthEndpoint: req.HealthEndpoint,
		Labels:         req.Labels,
	}

	result, err := h.svc.Register(r.Context(), n)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, result)
}

func (h *CollectorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	n, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, n)
}

func (h *CollectorHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	items, err := h.svc.List(r.Context(), q.Get("zone_id"), q.Get("node_type"), q.Get("status"))
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, items)
}

func (h *CollectorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name    string `json:"name"`
		Address string `json:"address"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	updates := &model.CollectorNode{
		Name:    req.Name,
		Address: req.Address,
		Status:  req.Status,
	}

	result, err := h.svc.Update(r.Context(), id, updates)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, result)
}

func (h *CollectorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *CollectorHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Heartbeat(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "ok"})
}
