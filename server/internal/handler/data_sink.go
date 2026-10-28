package handler

import (
	"dcs/internal/model"
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type DataSinkHandler struct {
	svc *service.DataSinkService
}

func NewDataSinkHandler(svc *service.DataSinkService) *DataSinkHandler {
	return &DataSinkHandler{svc: svc}
}

func (h *DataSinkHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *DataSinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID   string `json:"zone_id"`
		Name     string `json:"name"`
		SinkType string `json:"sink_type"`
		Protocol string `json:"protocol"`
		Config   string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	ds := &model.DataSink{
		ZoneID:   req.ZoneID,
		Name:     req.Name,
		SinkType: req.SinkType,
		Protocol: req.Protocol,
		Config:   req.Config,
	}

	result, err := h.svc.Create(r.Context(), ds)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, result)
}

func (h *DataSinkHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ds, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, ds)
}

func (h *DataSinkHandler) List(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	sinkType := r.URL.Query().Get("sink_type")

	items, err := h.svc.List(r.Context(), zoneID, sinkType)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, items)
}

func (h *DataSinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name     string `json:"name"`
		SinkType string `json:"sink_type"`
		Protocol string `json:"protocol"`
		Config   string `json:"config"`
		Status   string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	updates := &model.DataSink{
		Name:     req.Name,
		SinkType: req.SinkType,
		Protocol: req.Protocol,
		Config:   req.Config,
		Status:   req.Status,
	}

	result, err := h.svc.Update(r.Context(), id, updates)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, result)
}

func (h *DataSinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}
