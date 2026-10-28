package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type MetricDefinitionHandler struct {
	svc *service.MetricDefinitionService
}

func NewMetricDefinitionHandler(svc *service.MetricDefinitionService) *MetricDefinitionHandler {
	return &MetricDefinitionHandler{svc: svc}
}

func (h *MetricDefinitionHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/queries", h.CreateQuery)
	r.Get("/{id}/queries", h.ListQueries)
	r.Delete("/queries/{queryId}", h.DeleteQuery)
	return r
}

func (h *MetricDefinitionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name           string  `json:"name"`
		Category       string  `json:"category"`
		MetricType     *string `json:"metric_type"`
		Help           *string `json:"help"`
		Unit           *string `json:"unit"`
		PromqlTemplate *string `json:"promql_template"`
		Variables      *string `json:"variables"`
		Labels         *string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.Create(r.Context(), service.CreateMetricRequest{
		Name:           req.Name,
		Category:       req.Category,
		MetricType:     req.MetricType,
		Help:           req.Help,
		Unit:           req.Unit,
		PromqlTemplate: req.PromqlTemplate,
		Variables:      req.Variables,
		Labels:         req.Labels,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, m)
}

func (h *MetricDefinitionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	m, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, m)
}

func (h *MetricDefinitionHandler) List(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	metricType := r.URL.Query().Get("metric_type")
	keyword := r.URL.Query().Get("keyword")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), category, metricType, keyword, page, pageSize)
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

func (h *MetricDefinitionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name           string  `json:"name"`
		Category       string  `json:"category"`
		MetricType     *string `json:"metric_type"`
		Help           *string `json:"help"`
		Unit           *string `json:"unit"`
		PromqlTemplate *string `json:"promql_template"`
		Variables      *string `json:"variables"`
		Labels         *string `json:"labels"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	m, err := h.svc.Update(r.Context(), id, service.UpdateMetricRequest{
		Name:           req.Name,
		Category:       req.Category,
		MetricType:     req.MetricType,
		Help:           req.Help,
		Unit:           req.Unit,
		PromqlTemplate: req.PromqlTemplate,
		Variables:      req.Variables,
		Labels:         req.Labels,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, m)
}

func (h *MetricDefinitionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *MetricDefinitionHandler) CreateQuery(w http.ResponseWriter, r *http.Request) {
	metricID := chi.URLParam(r, "id")
	var req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		Promql      string  `json:"promql"`
		Variables   *string `json:"variables"`
		Category    *string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	q, err := h.svc.CreateQuery(r.Context(), metricID, service.CreateQueryRequest{
		Name:        req.Name,
		Description: req.Description,
		Promql:      req.Promql,
		Variables:   req.Variables,
		Category:    req.Category,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, q)
}

func (h *MetricDefinitionHandler) ListQueries(w http.ResponseWriter, r *http.Request) {
	metricID := chi.URLParam(r, "id")
	queries, err := h.svc.ListQueries(r.Context(), metricID)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, queries)
}

func (h *MetricDefinitionHandler) DeleteQuery(w http.ResponseWriter, r *http.Request) {
	queryID := chi.URLParam(r, "queryId")
	if err := h.svc.DeleteQuery(r.Context(), queryID); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}
