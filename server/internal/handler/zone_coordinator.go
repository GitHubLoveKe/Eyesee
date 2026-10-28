package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ZoneCoordinatorHandler struct {
	svc *service.ZoneCoordinatorService
}

func NewZoneCoordinatorHandler(svc *service.ZoneCoordinatorService) *ZoneCoordinatorHandler {
	return &ZoneCoordinatorHandler{svc: svc}
}

func (h *ZoneCoordinatorHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/heartbeat", h.Heartbeat)
	return r
}

func (h *ZoneCoordinatorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID     string `json:"zone_id"`
		Name       string `json:"name"`
		Address    string `json:"address"`
		GossipAddr string `json:"gossip_addr"`
		Role       string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	c, err := h.svc.Create(r.Context(), service.CreateZoneCoordinatorRequest{
		ZoneID:     req.ZoneID,
		Name:       req.Name,
		Address:    req.Address,
		GossipAddr: req.GossipAddr,
		Role:       req.Role,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, c)
}

func (h *ZoneCoordinatorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, c)
}

func (h *ZoneCoordinatorHandler) List(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	items, err := h.svc.List(r.Context(), zoneID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, items)
}

func (h *ZoneCoordinatorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name       string `json:"name"`
		Address    string `json:"address"`
		GossipAddr string `json:"gossip_addr"`
		Role       string `json:"role"`
		Status     string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	c, err := h.svc.Update(r.Context(), id, service.UpdateZoneCoordinatorRequest{
		Name:       req.Name,
		Address:    req.Address,
		GossipAddr: req.GossipAddr,
		Role:       req.Role,
		Status:     req.Status,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, c)
}

func (h *ZoneCoordinatorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *ZoneCoordinatorHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Heartbeat(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "ok"})
}
