package handler

import (
	"dcs/internal/repository"
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type AlertConvergenceHandler struct {
	svc *service.AlertConvergenceService
}

func NewAlertConvergenceHandler(svc *service.AlertConvergenceService) *AlertConvergenceHandler {
	return &AlertConvergenceHandler{svc: svc}
}

func (h *AlertConvergenceHandler) EventRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListEvents)
	r.Post("/", h.CreateEvent)
	r.Get("/statistics", h.GetStatistics)
	r.Get("/{id}", h.GetEvent)
	r.Post("/{id}/acknowledge", h.AcknowledgeEvent)
	return r
}

func (h *AlertConvergenceHandler) SilenceRoutes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.ListSilences)
	r.Post("/", h.CreateSilence)
	r.Post("/{id}/revoke", h.RevokeSilence)
	return r
}

func (h *AlertConvergenceHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AlertName   string  `json:"alert_name"`
		ZoneID      *string `json:"zone_id"`
		InstanceID  *string `json:"instance_id"`
		Severity    *string `json:"severity"`
		Labels      string  `json:"labels"`
		Annotations string  `json:"annotations"`
		Fingerprint *string `json:"fingerprint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	e, err := h.svc.CreateEvent(r.Context(), service.CreateAlertEventRequest{
		AlertName:   req.AlertName,
		ZoneID:      req.ZoneID,
		InstanceID:  req.InstanceID,
		Severity:    req.Severity,
		Labels:      req.Labels,
		Annotations: req.Annotations,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, e)
}

func (h *AlertConvergenceHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid event_id")
		return
	}
	e, err := h.svc.GetEvent(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, e)
}

func (h *AlertConvergenceHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	filter := repository.AlertEventFilter{
		ZoneID:      r.URL.Query().Get("zone_id"),
		AlertName:   r.URL.Query().Get("alert_name"),
		InstanceID:  r.URL.Query().Get("instance_id"),
		Fingerprint: r.URL.Query().Get("fingerprint"),
		Offset:      (page - 1) * pageSize,
		Limit:       pageSize,
	}

	if v := r.URL.Query().Get("severity"); v != "" {
		filter.Severity = strings.Split(v, ",")
	}
	if v := r.URL.Query().Get("status"); v != "" {
		filter.Status = strings.Split(v, ",")
	}

	items, total, err := h.svc.ListEvents(r.Context(), filter)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OKPage(w, items, total, page, pageSize)
}

func (h *AlertConvergenceHandler) AcknowledgeEvent(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "invalid event_id")
		return
	}
	var req struct {
		UserID string `json:"user_id"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ack, err := h.svc.AcknowledgeEvent(r.Context(), id, req.UserID, req.Note)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, ack)
}

func (h *AlertConvergenceHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "24h"
	}

	stats, err := h.svc.GetStatistics(r.Context(), zoneID, period)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, stats)
}

func (h *AlertConvergenceHandler) CreateSilence(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID   *string `json:"zone_id"`
		Matchers string  `json:"matchers"`
		Duration string  `json:"duration"`
		Reason   string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	silence, err := h.svc.CreateSilence(r.Context(), service.CreateSilenceRequest{
		ZoneID:    req.ZoneID,
		Matchers:  req.Matchers,
		Duration:  req.Duration,
		Reason:    req.Reason,
		CreatedBy: "admin",
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, silence)
}

func (h *AlertConvergenceHandler) ListSilences(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	status := r.URL.Query().Get("status")

	items, err := h.svc.ListSilences(r.Context(), zoneID, status)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, items)
}

func (h *AlertConvergenceHandler) RevokeSilence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.RevokeSilence(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "revoked"})
}
