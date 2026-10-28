package handler

import (
	"dcs/internal/model"
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TaskTypeHandler struct {
	svc *service.TaskTypeService
}

func NewTaskTypeHandler(svc *service.TaskTypeService) *TaskTypeHandler {
	return &TaskTypeHandler{svc: svc}
}

func (h *TaskTypeHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *TaskTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		SourceType      string `json:"source_type"`
		Collector       string `json:"collector"`
		FormSchema      string `json:"form_schema"`
		DetailPanels    string `json:"detail_panels"`
		DeclaredMetrics string `json:"declared_metrics"`
		ConfigTemplate  string `json:"config_template"`
		IsBuiltin       bool   `json:"is_builtin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	tt := &model.TaskType{
		Name:            req.Name,
		Description:     req.Description,
		SourceType:      req.SourceType,
		Collector:       req.Collector,
		FormSchema:      req.FormSchema,
		DetailPanels:    req.DetailPanels,
		DeclaredMetrics: req.DeclaredMetrics,
		ConfigTemplate:  req.ConfigTemplate,
		IsBuiltin:       req.IsBuiltin,
	}

	result, err := h.svc.Create(r.Context(), tt)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, result)
}

func (h *TaskTypeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tt, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, tt)
}

func (h *TaskTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.List(r.Context())
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, items)
}

func (h *TaskTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		SourceType      string `json:"source_type"`
		Collector       string `json:"collector"`
		FormSchema      string `json:"form_schema"`
		DetailPanels    string `json:"detail_panels"`
		DeclaredMetrics string `json:"declared_metrics"`
		ConfigTemplate  string `json:"config_template"`
		Version         int    `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	updates := &model.TaskType{
		Name:            req.Name,
		Description:     req.Description,
		SourceType:      req.SourceType,
		Collector:       req.Collector,
		FormSchema:      req.FormSchema,
		DetailPanels:    req.DetailPanels,
		DeclaredMetrics: req.DeclaredMetrics,
		ConfigTemplate:  req.ConfigTemplate,
		Version:         req.Version,
	}

	result, err := h.svc.Update(r.Context(), id, updates)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, result)
}

func (h *TaskTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}
