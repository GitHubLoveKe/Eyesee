package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServiceDiscoveryHandler struct {
	svc *service.ServiceDiscoveryService
}

func NewServiceDiscoveryHandler(svc *service.ServiceDiscoveryService) *ServiceDiscoveryHandler {
	return &ServiceDiscoveryHandler{svc: svc}
}

func (h *ServiceDiscoveryHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/targets", h.GetTargets)
	return r
}

func (h *ServiceDiscoveryHandler) GetTargets(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	nodeID := r.URL.Query().Get("node_id")

	if zoneID == "" || nodeID == "" {
		response.BadRequest(w, "zone_id and node_id are required")
		return
	}

	targets, err := h.svc.GetTargets(r.Context(), zoneID, nodeID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	if targets == nil {
		targets = []service.SDTarget{}
	}
	response.OK(w, targets)
}
