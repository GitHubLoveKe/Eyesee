package handler

import (
	"dcs/internal/service"
	"dcs/pkg/response"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type CredentialHandler struct {
	svc *service.CredentialService
}

func NewCredentialHandler(svc *service.CredentialService) *CredentialHandler {
	return &CredentialHandler{svc: svc}
}

func (h *CredentialHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *CredentialHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ZoneID    string     `json:"zone_id"`
		Name      string     `json:"name"`
		CredType  string     `json:"cred_type"`
		Data      string     `json:"data"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	c, err := h.svc.Create(r.Context(), service.CreateCredentialRequest{
		ZoneID:    req.ZoneID,
		Name:      req.Name,
		CredType:  req.CredType,
		Data:      req.Data,
		ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.Created(w, c)
}

func (h *CredentialHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, c)
}

func (h *CredentialHandler) List(w http.ResponseWriter, r *http.Request) {
	zoneID := r.URL.Query().Get("zone_id")
	credType := r.URL.Query().Get("cred_type")

	items, err := h.svc.List(r.Context(), zoneID, credType)
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, items)
}

func (h *CredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Name      string     `json:"name"`
		CredType  string     `json:"cred_type"`
		Data      string     `json:"data"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	c, err := h.svc.Update(r.Context(), id, service.UpdateCredentialRequest{
		Name:      req.Name,
		CredType:  req.CredType,
		Data:      req.Data,
		ExpiresAt: req.ExpiresAt,
	})
	if err != nil {
		handleAppError(w, err)
		return
	}
	response.OK(w, c)
}

func (h *CredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		handleAppError(w, err)
		return
	}
	response.NoContent(w)
}
