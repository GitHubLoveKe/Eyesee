package handler

import (
	"dcs/internal/middleware"
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type BusinessSystemHandler struct {
	svc *service.BusinessSystemService
}

func NewBusinessSystemHandler(svc *service.BusinessSystemService) *BusinessSystemHandler {
	return &BusinessSystemHandler{svc: svc}
}

func (h *BusinessSystemHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/my", h.ListMy)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/suspend", h.Suspend)
	r.Post("/{id}/archive", h.Archive)

	r.Post("/{id}/members", h.AddMember)
	r.Get("/{id}/members", h.ListMembers)
	r.Delete("/{id}/members/{userId}", h.RemoveMember)
	r.Put("/{id}/members/{userId}", h.UpdateMemberRole)

	return r
}

func (h *BusinessSystemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	bs, err := h.svc.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.Created(w, bs)
}

func (h *BusinessSystemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	bs, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, bs)
}

func (h *BusinessSystemHandler) List(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	items, total, err := h.svc.List(r.Context(), keyword, page, pageSize)
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

func (h *BusinessSystemHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	bs, err := h.svc.Update(r.Context(), id, req.Name, req.Description, req.Status)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, bs)
}

func (h *BusinessSystemHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}

	response.NoContent(w)
}

func (h *BusinessSystemHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.AddMember(r.Context(), id, req.UserID, req.Role); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "added"})
}

func (h *BusinessSystemHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	members, err := h.svc.ListMembers(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, members)
}

func (h *BusinessSystemHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")

	if err := h.svc.RemoveMember(r.Context(), id, userID); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "removed"})
}

func (h *BusinessSystemHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.svc.UpdateMemberRole(r.Context(), id, userID, req.Role); err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, map[string]string{"status": "updated"})
}

func (h *BusinessSystemHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == "" {
		response.Unauthorized(w, "missing user context")
		return
	}

	systems, err := h.svc.ListByUser(r.Context(), userID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response.OK(w, systems)
}

func (h *BusinessSystemHandler) Suspend(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	bs, err := h.svc.Suspend(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, bs)
}

func (h *BusinessSystemHandler) Archive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	bs, err := h.svc.Archive(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, bs)
}
