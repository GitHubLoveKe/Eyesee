package handler

import (
	"dcs/internal/model"
	"dcs/internal/repository"
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CollectionTaskHandler struct {
	svc *service.CollectionTaskService
}

func NewCollectionTaskHandler(svc *service.CollectionTaskService) *CollectionTaskHandler {
	return &CollectionTaskHandler{svc: svc}
}

func (h *CollectionTaskHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/enable", h.Enable)
	r.Post("/{id}/disable", h.Disable)
	r.Post("/batch/enable", h.BatchEnable)
	r.Post("/batch/disable", h.BatchDisable)
	r.Post("/batch/delete", h.BatchDelete)

	return r
}

func (h *CollectionTaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InstanceID string  `json:"instance_id"`
		TypeID     string  `json:"type_id"`
		Module     string  `json:"module"`
		DataSinkID *string `json:"data_sink_id"`
		Name       string  `json:"name"`
		Config     string  `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ct := &model.CollectionTask{
		InstanceID: req.InstanceID,
		TypeID:     req.TypeID,
		Module:     req.Module,
		DataSinkID: req.DataSinkID,
		Name:       req.Name,
		Config:     req.Config,
	}

	result, err := h.svc.Create(r.Context(), ct)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, result)
}

func (h *CollectionTaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ct, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, ct)
}

func (h *CollectionTaskHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var enabled *bool
	if e := q.Get("enabled"); e != "" {
		b := e == "true"
		enabled = &b
	}

	filter := repository.CollectionTaskFilter{
		InstanceID: q.Get("instance_id"),
		TypeID:     q.Get("type_id"),
		NodeID:     q.Get("node_id"),
		Status:     q.Get("status"),
		Enabled:    enabled,
		Keyword:    q.Get("keyword"),
		Offset:     (page - 1) * pageSize,
		Limit:      pageSize,
	}

	items, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OKPage(w, items, total, page, pageSize)
}

func (h *CollectionTaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name    string `json:"name"`
		Config  string `json:"config"`
		Enabled bool   `json:"enabled"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	updates := &model.CollectionTask{
		Name:    req.Name,
		Config:  req.Config,
		Enabled: req.Enabled,
		Status:  req.Status,
	}

	result, err := h.svc.Update(r.Context(), id, updates)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, result)
}

func (h *CollectionTaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *CollectionTaskHandler) Enable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Enable(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "enabled"})
}

func (h *CollectionTaskHandler) Disable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Disable(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "disabled"})
}

func (h *CollectionTaskHandler) BatchEnable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(w, "ids array cannot be empty")
		return
	}
	if len(req.IDs) > 100 {
		response.BadRequest(w, "batch size cannot exceed 100")
		return
	}
	var succeeded, failed int
	for _, id := range req.IDs {
		if err := h.svc.Enable(r.Context(), id); err != nil {
			failed++
		} else {
			succeeded++
		}
	}
	response.OK(w, map[string]int{"succeeded": succeeded, "failed": failed})
}

func (h *CollectionTaskHandler) BatchDisable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(w, "ids array cannot be empty")
		return
	}
	if len(req.IDs) > 100 {
		response.BadRequest(w, "batch size cannot exceed 100")
		return
	}
	var succeeded, failed int
	for _, id := range req.IDs {
		if err := h.svc.Disable(r.Context(), id); err != nil {
			failed++
		} else {
			succeeded++
		}
	}
	response.OK(w, map[string]int{"succeeded": succeeded, "failed": failed})
}

func (h *CollectionTaskHandler) BatchDelete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}
	if len(req.IDs) == 0 {
		response.BadRequest(w, "ids array cannot be empty")
		return
	}
	if len(req.IDs) > 100 {
		response.BadRequest(w, "batch size cannot exceed 100")
		return
	}
	var succeeded, failed int
	for _, id := range req.IDs {
		if err := h.svc.Delete(r.Context(), id); err != nil {
			failed++
		} else {
			succeeded++
		}
	}
	response.OK(w, map[string]int{"succeeded": succeeded, "failed": failed})
}
