package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type BusinessOverviewHandler struct {
	svc *service.BusinessOverviewService
}

func NewBusinessOverviewHandler(svc *service.BusinessOverviewService) *BusinessOverviewHandler {
	return &BusinessOverviewHandler{svc: svc}
}

func (h *BusinessOverviewHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/{systemId}", h.GetOverview)
	return r
}

func (h *BusinessOverviewHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	systemID := chi.URLParam(r, "systemId")
	overview, err := h.svc.GetOverview(r.Context(), systemID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, overview)
}
