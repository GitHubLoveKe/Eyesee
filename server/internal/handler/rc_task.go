package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type RCTaskHandler struct {
	svc *service.RCTaskService
}

func NewRCTaskHandler(svc *service.RCTaskService) *RCTaskHandler {
	return &RCTaskHandler{svc: svc}
}

func (h *RCTaskHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/rules", h.CreateRule)
	r.Get("/{id}/rules", h.ListRules)
	r.Put("/rules/{ruleId}", h.UpdateRule)
	r.Delete("/rules/{ruleId}", h.DeleteRule)
	return r
}

func (h *RCTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID       string  `json:"zone_id"`
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		StrategyID   *string `json:"strategy_id"`
		QueryAddress *string `json:"query_address"`
		RuleGroup    *string `json:"rule_group"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	t, err := h.svc.Create(r.Context(), service.CreateRCTaskRequest{
		ZoneID:       req.ZoneID,
		Name:         req.Name,
		Description:  req.Description,
		StrategyID:   req.StrategyID,
		QueryAddress: req.QueryAddress,
		RuleGroup:    req.RuleGroup,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, t)
}

func (h *RCTaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	t, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, t)
}

func (h *RCTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	keyword := r.URL.Query().Get("keyword")
	var enabled *bool
	if v := r.URL.Query().Get("enabled"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			enabled = &b
		}
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), zoneID, enabled, keyword, page, pageSize)
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

func (h *RCTaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name         string  `json:"name"`
		Description  string  `json:"description"`
		QueryAddress *string `json:"query_address"`
		Enabled      *bool   `json:"enabled"`
		Status       string  `json:"status"`
		RuleGroup    *string `json:"rule_group"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	t, err := h.svc.Update(r.Context(), id, service.UpdateRCTaskRequest{
		Name:         req.Name,
		Description:  req.Description,
		QueryAddress: req.QueryAddress,
		Enabled:      req.Enabled,
		Status:       req.Status,
		RuleGroup:    req.RuleGroup,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, t)
}

func (h *RCTaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *RCTaskHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	var req struct {
		Module      string `json:"module"`
		AlertName   string `json:"alert_name"`
		Expr        string `json:"expr"`
		Duration    string `json:"duration"`
		Severity    string `json:"severity"`
		Labels      string `json:"labels"`
		Annotations string `json:"annotations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	rule, err := h.svc.CreateRule(r.Context(), taskID, service.CreateRCRuleRequest{
		Module:      req.Module,
		AlertName:   req.AlertName,
		Expr:        req.Expr,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, rule)
}

func (h *RCTaskHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	rules, err := h.svc.ListRules(r.Context(), taskID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, rules)
}

func (h *RCTaskHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	var req struct {
		AlertName   string `json:"alert_name"`
		Expr        string `json:"expr"`
		Duration    string `json:"duration"`
		Severity    string `json:"severity"`
		Labels      string `json:"labels"`
		Annotations string `json:"annotations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.UpdateRule(r.Context(), ruleID, service.UpdateRCRuleRequest{
		AlertName:   req.AlertName,
		Expr:        req.Expr,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	}); err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, map[string]string{"status": "updated"})
}

func (h *RCTaskHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleId")
	if err := h.svc.DeleteRule(r.Context(), ruleID); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}
