package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ResourceTypeHandler struct {
	svc *service.ResourceSystemService
}

func NewResourceTypeHandler(svc *service.ResourceSystemService) *ResourceTypeHandler {
	return &ResourceTypeHandler{svc: svc}
}

func (h *ResourceTypeHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/tree", h.GetTree)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *ResourceTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ParentID    *string `json:"parent_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Icon        string  `json:"icon"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	rt, err := h.svc.Create(r.Context(), req.ParentID, req.Name, req.Description, req.Icon, req.SortOrder)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, rt)
}

func (h *ResourceTypeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	rt, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, rt)
}

func (h *ResourceTypeHandler) GetTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.svc.GetTree(r.Context())
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]interface{}{"tree": tree})
}

func (h *ResourceTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	parentID := r.URL.Query().Get("parent_id")
	var pid *string
	if parentID != "" {
		pid = &parentID
	}

	items, err := h.svc.List(r.Context(), pid)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, items)
}

func (h *ResourceTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		ParentID    *string `json:"parent_id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Icon        string  `json:"icon"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	rt, err := h.svc.Update(r.Context(), id, req.ParentID, req.Name, req.Description, req.Icon, req.SortOrder)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, rt)
}

func (h *ResourceTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}
