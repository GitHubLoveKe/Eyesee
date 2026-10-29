package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ServiceDiscoveryConfigHandler struct {
	svc *service.ServiceDiscoveryConfigService
}

func NewServiceDiscoveryConfigHandler(svc *service.ServiceDiscoveryConfigService) *ServiceDiscoveryConfigHandler {
	return &ServiceDiscoveryConfigHandler{svc: svc}
}

func (h *ServiceDiscoveryConfigHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *ServiceDiscoveryConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID          string `json:"zone_id"`
		Name            string `json:"name"`
		Protocol        string `json:"protocol"`
		Config          string `json:"config"`
		RefreshInterval string `json:"refresh_interval"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	sd, err := h.svc.Create(r.Context(), service.CreateSDConfigRequest{
		ZoneID:          req.ZoneID,
		Name:            req.Name,
		Protocol:        req.Protocol,
		Config:          req.Config,
		RefreshInterval: req.RefreshInterval,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, sd)
}

func (h *ServiceDiscoveryConfigHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	sd, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, sd)
}

func (h *ServiceDiscoveryConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), zoneID, page, pageSize)
	if err != nil {
		handleAppError(w, err)
		return
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	response.OKPage(w, items, total, page, pageSize)
}

func (h *ServiceDiscoveryConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name            string `json:"name"`
		Protocol        string `json:"protocol"`
		Config          string `json:"config"`
		RefreshInterval string `json:"refresh_interval"`
		Enabled         *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	sd, err := h.svc.Update(r.Context(), id, service.UpdateSDConfigRequest{
		Name:            req.Name,
		Protocol:        req.Protocol,
		Config:          req.Config,
		RefreshInterval: req.RefreshInterval,
		Enabled:         req.Enabled,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, sd)
}

func (h *ServiceDiscoveryConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}
