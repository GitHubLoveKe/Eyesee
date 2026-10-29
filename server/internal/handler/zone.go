package handler

import (
	"dcs/internal/service"
	"dcs/pkg/errors"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func handleAppError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		response.Error(w, appErr.HTTPStatus, appErr.Code, appErr.Message)
		return
	}
	response.InternalError(w, "internal server error")
}

type ZoneHandler struct {
	svc *service.ZoneService
}

func NewZoneHandler(svc *service.ZoneService) *ZoneHandler {
	return &ZoneHandler{svc: svc}
}

func (h *ZoneHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)

	return r
}

func (h *ZoneHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		CollectorType string `json:"collector_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	z, err := h.svc.Create(r.Context(), req.Name, req.Description, req.CollectorType)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, z)
}

func (h *ZoneHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	z, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, z)
}

func (h *ZoneHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), status, page, pageSize)
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

func (h *ZoneHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		CollectorType string `json:"collector_type"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	z, err := h.svc.Update(r.Context(), id, req.Name, req.Description, req.CollectorType, req.Status)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, z)
}

func (h *ZoneHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *ZoneHandler) CreateSegment(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "id")
	var req struct {
		CIDR        string `json:"cidr"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	seg, err := h.svc.CreateSegment(r.Context(), zoneID, req.CIDR, req.Name, req.Description)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, seg)
}

func (h *ZoneHandler) ListSegments(w http.ResponseWriter, r *http.Request) {
	zoneID := chi.URLParam(r, "id")
	segs, err := h.svc.ListSegments(r.Context(), zoneID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, segs)
}

func (h *ZoneHandler) UpdateSegment(w http.ResponseWriter, r *http.Request) {
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

	if err := h.svc.UpdateSegment(r.Context(), segmentID, req.CIDR, req.Name, req.Description); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "updated"})
}

func (h *ZoneHandler) DeleteSegment(w http.ResponseWriter, r *http.Request) {
	segmentID, _ := strconv.ParseInt(chi.URLParam(r, "segmentId"), 10, 64)
	if err := h.svc.DeleteSegment(r.Context(), segmentID); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}
