package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ScheduledTaskHandler struct {
	svc *service.ScheduledTaskService
}

func NewScheduledTaskHandler(svc *service.ScheduledTaskService) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{svc: svc}
}

func (h *ScheduledTaskHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/enable", h.Enable)
	r.Post("/{id}/disable", h.Disable)
	r.Post("/{id}/trigger", h.Trigger)
	r.Get("/{id}/executions", h.ListExecutions)
	return r
}

func (h *ScheduledTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BusinessSysID *string `json:"business_system_id"`
		Name          string  `json:"name"`
		TaskType      string  `json:"task_type"`
		CronExpr      string  `json:"cron_expr"`
		Config        string  `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	t, err := h.svc.Create(r.Context(), service.CreateScheduledTaskRequest{
		BusinessSysID: req.BusinessSysID,
		Name:          req.Name,
		TaskType:      req.TaskType,
		CronExpr:      req.CronExpr,
		Config:        req.Config,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, t)
}

func (h *ScheduledTaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	t, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, t)
}

func (h *ScheduledTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	taskType := r.URL.Query().Get("task_type")
	status := r.URL.Query().Get("status")
	keyword := r.URL.Query().Get("keyword")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), taskType, status, nil, keyword, page, pageSize)
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

func (h *ScheduledTaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name     string `json:"name"`
		TaskType string `json:"task_type"`
		CronExpr string `json:"cron_expr"`
		Config   string `json:"config"`
		Enabled  *bool  `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	t, err := h.svc.Update(r.Context(), id, service.UpdateScheduledTaskRequest{
		Name:     req.Name,
		TaskType: req.TaskType,
		CronExpr: req.CronExpr,
		Config:   req.Config,
		Enabled:  req.Enabled,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, t)
}

func (h *ScheduledTaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *ScheduledTaskHandler) Enable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	t, err := h.svc.Enable(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, t)
}

func (h *ScheduledTaskHandler) Disable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	t, err := h.svc.Disable(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, t)
}

func (h *ScheduledTaskHandler) ListExecutions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	executions, err := h.svc.ListExecutions(r.Context(), id, limit)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, executions)
}

func (h *ScheduledTaskHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	execution, err := h.svc.Trigger(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, execution)
}
