package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// EntityHandler handles entity-related HTTP requests.
type EntityHandler struct {
	service service.EntityService
}

// NewEntityHandler creates a new entity handler.
func NewEntityHandler(svc service.EntityService) *EntityHandler {
	return &EntityHandler{service: svc}
}

type createEntityRequest struct {
	Name   string            `json:"name"`
	Type   string            `json:"type"`
	Fields map[string]string `json:"fields"`
}

type updateEntityRequest struct {
	Name   *string            `json:"name,omitempty"`
	Type   *string            `json:"type,omitempty"`
	Fields *map[string]string `json:"fields,omitempty"`
}

type linkEntityRequest struct {
	EntityID string `json:"entityId"`
}

type fromContactRequest struct {
	ContactUserID string `json:"contactUserId"`
}

// Create handles POST /entities
func (h *EntityHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	var req createEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	entity, err := h.service.Create(r.Context(), userID, service.CreateEntityInput{
		Name:   req.Name,
		Type:   req.Type,
		Fields: req.Fields,
	})
	if err != nil {
		handleEntityError(w, err)
		return
	}

	response.Created(w, entity)
}

// List handles GET /entities
func (h *EntityHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	entityType := r.URL.Query().Get("type")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 20
	}

	items, total, err := h.service.List(r.Context(), userID, entityType, limit, offset)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	if items == nil {
		items = []*model.EntityListItem{}
	}

	response.OK(w, map[string]any{
		"data":   items,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetByID handles GET /entities/{id}
func (h *EntityHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format entity ID tidak valid"))
		return
	}

	entity, err := h.service.GetByID(r.Context(), entityID, userID)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	response.OK(w, entity)
}

// Update handles PUT /entities/{id}
func (h *EntityHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format entity ID tidak valid"))
		return
	}

	var req updateEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	entity, err := h.service.Update(r.Context(), entityID, userID, service.UpdateEntityInput{
		Name:   req.Name,
		Type:   req.Type,
		Fields: req.Fields,
	})
	if err != nil {
		handleEntityError(w, err)
		return
	}

	response.OK(w, entity)
}

// Delete handles DELETE /entities/{id}
func (h *EntityHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format entity ID tidak valid"))
		return
	}

	if err := h.service.Delete(r.Context(), entityID, userID); err != nil {
		handleEntityError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "entity berhasil dihapus"})
}

// Search handles GET /entities/search?q=...
func (h *EntityHandler) Search(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	query := r.URL.Query().Get("q")

	entities, err := h.service.Search(r.Context(), userID, query)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	if entities == nil {
		entities = []*model.Entity{}
	}

	response.OK(w, entities)
}

// ListTypes handles GET /entities/types
func (h *EntityHandler) ListTypes(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	types, err := h.service.ListTypes(r.Context(), userID)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	if types == nil {
		types = []string{}
	}

	response.OK(w, types)
}

// ListDocuments handles GET /entities/{id}/documents
func (h *EntityHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	entityID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format entity ID tidak valid"))
		return
	}

	docs, err := h.service.GetEntityDocuments(r.Context(), entityID)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	if docs == nil {
		docs = []*model.Document{}
	}

	response.OK(w, docs)
}

// LinkToDocument handles POST /documents/{id}/entities
func (h *EntityHandler) LinkToDocument(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	var req linkEntityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		response.Error(w, apperror.BadRequest("format entity ID tidak valid"))
		return
	}

	if err := h.service.LinkToDocument(r.Context(), entityID, docID, userID); err != nil {
		handleEntityError(w, err)
		return
	}

	response.Created(w, map[string]string{"message": "entity berhasil ditautkan"})
}

// UnlinkFromDocument handles DELETE /documents/{id}/entities/{entityId}
func (h *EntityHandler) UnlinkFromDocument(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	entityID, err := uuid.Parse(chi.URLParam(r, "entityId"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format entity ID tidak valid"))
		return
	}

	if err := h.service.UnlinkFromDocument(r.Context(), entityID, docID, userID); err != nil {
		handleEntityError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "entity berhasil dilepas"})
}

// GetDocumentEntities handles GET /documents/{id}/entities
func (h *EntityHandler) GetDocumentEntities(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	entities, err := h.service.GetDocumentEntities(r.Context(), docID)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	if entities == nil {
		entities = []*model.Entity{}
	}

	response.OK(w, entities)
}

// CreateFromContact handles POST /entities/from-contact
func (h *EntityHandler) CreateFromContact(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	var req fromContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	contactUserID, err := uuid.Parse(req.ContactUserID)
	if err != nil {
		response.Error(w, apperror.BadRequest("format contact user ID tidak valid"))
		return
	}

	entity, err := h.service.CreateFromContact(r.Context(), contactUserID, userID)
	if err != nil {
		handleEntityError(w, err)
		return
	}

	response.Created(w, entity)
}

func handleEntityError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		response.Error(w, appErr)
		return
	}
	response.Error(w, apperror.Internal(fmt.Errorf("kesalahan internal server")))
}
