package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type StatusCacheHandler struct {
	svc *service.StatusCacheService
}

func NewStatusCacheHandler(svc *service.StatusCacheService) *StatusCacheHandler {
	return &StatusCacheHandler{svc: svc}
}

func (h *StatusCacheHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/global", h.GetGlobalSummary)
	r.Get("/zones/{zoneId}", h.GetZoneOverview)
	r.Post("/refresh", h.Refresh)

	return r
}

func (h *StatusCacheHandler) GetGlobalSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.svc.GetGlobalSummary(r.Context())
	if err != nil {
		handleAppError(w, err)
		return
	}

	if summary == nil {
		summary = &service.GlobalSummary{}
	}

	response.OK(w, summary)
}

func (h *StatusCacheHandler) GetZoneOverview(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "zoneId")
	overview, err := h.svc.GetZoneOverview(r.Context(), zoneID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	if overview == nil {
		response.NotFound(w, "zone overview not found")
		return
	}

	response.OK(w, overview)
}

func (h *StatusCacheHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	response.OK(w, map[string]string{"status": "refresh triggered"})
}
