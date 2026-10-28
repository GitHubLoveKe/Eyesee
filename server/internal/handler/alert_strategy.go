package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type AlertStrategyHandler struct {
	svc *service.AlertStrategyService
}

func NewAlertStrategyHandler(svc *service.AlertStrategyService) *AlertStrategyHandler {
	return &AlertStrategyHandler{svc: svc}
}

func (h *AlertStrategyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/instances", h.LinkInstance)
	r.Get("/{id}/instances", h.ListLinkedInstances)
	r.Delete("/{id}/instances/{instanceId}", h.UnlinkInstance)
	return r
}

func (h *AlertStrategyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		MetricID     string  `json:"metric_id"`
		ExprTemplate string  `json:"expr_template"`
		Duration     string  `json:"duration"`
		Severity     string  `json:"severity"`
		ScopeType    string  `json:"scope_type"`
		ScopeValue   string  `json:"scope_value"`
		AlertGroupID *string `json:"alert_group_id"`
		ChannelID    *string `json:"channel_id"`
		Labels       string  `json:"labels"`
		Annotations  string  `json:"annotations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	s, err := h.svc.Create(r.Context(), service.CreateAlertStrategyRequest{
		Name:         req.Name,
		Description:  req.Description,
		MetricID:     req.MetricID,
		ExprTemplate: req.ExprTemplate,
		Duration:     req.Duration,
		Severity:     req.Severity,
		ScopeType:    req.ScopeType,
		ScopeValue:   req.ScopeValue,
		AlertGroupID: req.AlertGroupID,
		ChannelID:    req.ChannelID,
		Labels:       req.Labels,
		Annotations:  req.Annotations,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, s)
}

func (h *AlertStrategyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	s, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, s)
}

func (h *AlertStrategyHandler) List(w http.ResponseWriter, r *http.Request) {
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), severity, status, nil, keyword, page, pageSize)
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

func (h *AlertStrategyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		MetricID     string  `json:"metric_id"`
		ExprTemplate string  `json:"expr_template"`
		Duration     string  `json:"duration"`
		Severity     string  `json:"severity"`
		ScopeType    string  `json:"scope_type"`
		ScopeValue   string  `json:"scope_value"`
		AlertGroupID *string `json:"alert_group_id"`
		ChannelID    *string `json:"channel_id"`
		Labels       string  `json:"labels"`
		Annotations  string  `json:"annotations"`
		Enabled      *bool   `json:"enabled"`
		Status       string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	s, err := h.svc.Update(r.Context(), id, service.UpdateAlertStrategyRequest{
		Name:         req.Name,
		Description:  req.Description,
		MetricID:     req.MetricID,
		ExprTemplate: req.ExprTemplate,
		Duration:     req.Duration,
		Severity:     req.Severity,
		ScopeType:    req.ScopeType,
		ScopeValue:   req.ScopeValue,
		AlertGroupID: req.AlertGroupID,
		ChannelID:    req.ChannelID,
		Labels:       req.Labels,
		Annotations:  req.Annotations,
		Enabled:      req.Enabled,
		Status:       req.Status,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, s)
}

func (h *AlertStrategyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *AlertStrategyHandler) LinkInstance(w http.ResponseWriter, r *http.Request) {
	strategyID := chi.URLParam(r, "id")
	var req struct {
		InstanceID string `json:"instance_id"`
		ZoneID     string `json:"zone_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	si, err := h.svc.LinkInstance(r.Context(), strategyID, req.InstanceID, req.ZoneID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, si)
}

func (h *AlertStrategyHandler) ListLinkedInstances(w http.ResponseWriter, r *http.Request) {
	strategyID := chi.URLParam(r, "id")
	items, err := h.svc.ListLinkedInstances(r.Context(), strategyID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, items)
}

func (h *AlertStrategyHandler) UnlinkInstance(w http.ResponseWriter, r *http.Request) {
	strategyID := chi.URLParam(r, "id")
	instanceID := chi.URLParam(r, "instanceId")
	if err := h.svc.UnlinkInstance(r.Context(), strategyID, instanceID); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}
