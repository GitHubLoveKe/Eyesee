package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type NetworkSegmentHandler struct {
	svc *service.NetworkSegmentService
}

func NewNetworkSegmentHandler(svc *service.NetworkSegmentService) *NetworkSegmentHandler {
	return &NetworkSegmentHandler{svc: svc}
}

func (h *NetworkSegmentHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.ListByZone)
	r.Get("/{segmentId}", h.GetByID)
	r.Put("/{segmentId}", h.Update)
	r.Delete("/{segmentId}", h.Delete)

	return r
}

func (h *NetworkSegmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "zoneId")
	var req struct {
		CIDR        string `json:"cidr"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	seg, err := h.svc.Create(r.Context(), zoneID, req.CIDR, req.Name, req.Description)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, seg)
}

func (h *NetworkSegmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	segmentID, _ := strconv.ParseInt(chi.URLParam(r, "segmentId"), 10, 64)
	seg, err := h.svc.GetByID(r.Context(), segmentID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, seg)
}

func (h *NetworkSegmentHandler) ListByZone(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "zoneId")
	segs, err := h.svc.ListByZoneID(r.Context(), zoneID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, segs)
}

func (h *NetworkSegmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	segmentID, _ := strconv.ParseInt(chi.URLParam(r, "segmentId"), 10, 64)
	var req struct {
		CIDR        string `json:"cidr"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	seg, err := h.svc.Update(r.Context(), segmentID, req.CIDR, req.Name, req.Description)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, seg)
}

func (h *NetworkSegmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	segmentID, _ := strconv.ParseInt(chi.URLParam(r, "segmentId"), 10, 64)
	if err := h.svc.Delete(r.Context(), segmentID); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}
