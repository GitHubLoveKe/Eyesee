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

type InstanceHandler struct {
	svc *service.InstanceService
}

func NewInstanceHandler(svc *service.InstanceService) *InstanceHandler {
	return &InstanceHandler{svc: svc}
}

func (h *InstanceHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *InstanceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID         string  `json:"zone_id"`
		ResourceTypeID string  `json:"resource_type_id"`
		BusinessSysID  *string `json:"business_system_id"`
		Name           string  `json:"name"`
		Host           string  `json:"host"`
		Port           *int    `json:"port"`
		CredentialID   *string `json:"credential_id"`
		Labels         string  `json:"labels"`
		Extra          string  `json:"extra"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	inst := &model.Instance{
		ZoneID:         req.ZoneID,
		ResourceTypeID: req.ResourceTypeID,
		BusinessSysID:  req.BusinessSysID,
		Name:           req.Name,
		Host:           req.Host,
		Port:           req.Port,
		CredentialID:   req.CredentialID,
		Labels:         req.Labels,
		Extra:          req.Extra,
	}

	result, err := h.svc.Create(r.Context(), inst)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, result)
}

func (h *InstanceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inst, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, inst)
}

func (h *InstanceHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	filter := repository.InstanceFilter{
		ZoneID:         q.Get("zone_id"),
		ResourceTypeID: q.Get("resource_type_id"),
		BusinessSysID:  q.Get("business_system_id"),
		Status:         q.Get("status"),
		Keyword:        q.Get("keyword"),
		Offset:         (page - 1) * pageSize,
		Limit:          pageSize,
	}

	items, total, err := h.svc.List(r.Context(), filter)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OKPage(w, items, total, page, pageSize)
}

func (h *InstanceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name   string `json:"name"`
		Host   string `json:"host"`
		Port   *int   `json:"port"`
		Labels string `json:"labels"`
		Extra  string `json:"extra"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	updates := &model.Instance{
		Name:   req.Name,
		Host:   req.Host,
		Port:   req.Port,
		Labels: req.Labels,
		Extra:  req.Extra,
		Status: req.Status,
	}

	result, err := h.svc.Update(r.Context(), id, updates)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, result)
}

func (h *InstanceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}
