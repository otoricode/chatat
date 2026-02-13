package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// DocumentHandler handles document endpoints.
type DocumentHandler struct {
	documentService service.DocumentService
	blockService    service.BlockService
	templateService service.TemplateService
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(
	documentService service.DocumentService,
	blockService service.BlockService,
	templateService service.TemplateService,
) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
		blockService:    blockService,
		templateService: templateService,
	}
}

// -- Request types --

type createDocumentRequest struct {
	Title        string  `json:"title"`
	Icon         string  `json:"icon"`
	Cover        string  `json:"cover"`
	ChatID       *string `json:"chatId"`
	TopicID      *string `json:"topicId"`
	IsStandalone bool    `json:"isStandalone"`
	TemplateID   string  `json:"templateId"`
}

type updateDocumentRequest struct {
	Title       *string `json:"title"`
	Icon        *string `json:"icon"`
	Cover       *string `json:"cover"`
	RequireSigs *bool   `json:"requireSigs"`
}

type addBlockRequest struct {
	Type          string          `json:"type"`
	Content       string          `json:"content"`
	Position      int             `json:"position"`
	Checked       *bool           `json:"checked"`
	Rows          json.RawMessage `json:"rows"`
	Columns       json.RawMessage `json:"columns"`
	Language      string          `json:"language"`
	Emoji         string          `json:"emoji"`
	Color         string          `json:"color"`
	ParentBlockID *string         `json:"parentBlockId"`
}

type updateBlockRequest struct {
	Content  *string         `json:"content"`
	Checked  *bool           `json:"checked"`
	Rows     json.RawMessage `json:"rows"`
	Columns  json.RawMessage `json:"columns"`
	Language *string         `json:"language"`
	Emoji    *string         `json:"emoji"`
	Color    *string         `json:"color"`
}

type reorderBlocksRequest struct {
	BlockIDs []string `json:"blockIds"`
}

type batchBlockRequest struct {
	Operations []service.BlockOperation `json:"operations"`
}

type addCollaboratorRequest struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

type updateCollaboratorRoleRequest struct {
	Role string `json:"role"`
}

type addTagRequest struct {
	Tag string `json:"tag"`
}

// -- Document endpoints --

// Create handles POST /api/v1/documents
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	var req createDocumentRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	input := service.CreateDocumentInput{
		Title:        req.Title,
		Icon:         req.Icon,
		Cover:        req.Cover,
		OwnerID:      userID,
		IsStandalone: req.IsStandalone,
		TemplateID:   req.TemplateID,
	}

	if req.ChatID != nil {
		chatID, err := uuid.Parse(*req.ChatID)
		if err != nil {
			response.Error(w, apperror.BadRequest("format chatId tidak valid"))
			return
		}
		input.ChatID = &chatID
	}

	if req.TopicID != nil {
		topicID, err := uuid.Parse(*req.TopicID)
		if err != nil {
			response.Error(w, apperror.BadRequest("format topicId tidak valid"))
			return
		}
		input.TopicID = &topicID
	}

	doc, err := h.documentService.Create(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}

	response.Created(w, doc)
}

// GetByID handles GET /api/v1/documents/{id}
func (h *DocumentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
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

	doc, err := h.documentService.GetByID(r.Context(), docID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, doc)
}

// List handles GET /api/v1/documents
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	docs, err := h.documentService.ListAll(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, docs)
}

// ListByChat handles GET /api/v1/chats/{id}/documents
func (h *DocumentHandler) ListByChat(w http.ResponseWriter, r *http.Request) {
	chatID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format chat ID tidak valid"))
		return
	}

	docs, err := h.documentService.ListByContext(r.Context(), "chat", chatID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, docs)
}

// ListByTopic handles GET /api/v1/topics/{id}/documents
func (h *DocumentHandler) ListByTopic(w http.ResponseWriter, r *http.Request) {
	topicID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format topic ID tidak valid"))
		return
	}

	docs, err := h.documentService.ListByContext(r.Context(), "topic", topicID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, docs)
}

// Update handles PUT /api/v1/documents/{id}
func (h *DocumentHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	var req updateDocumentRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	doc, err := h.documentService.Update(r.Context(), docID, userID, model.UpdateDocumentInput{
		Title:       req.Title,
		Icon:        req.Icon,
		Cover:       req.Cover,
		RequireSigs: req.RequireSigs,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, doc)
}

// Delete handles DELETE /api/v1/documents/{id}
func (h *DocumentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	if err := h.documentService.Delete(r.Context(), docID, userID); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"deleted": true})
}

// Duplicate handles POST /api/v1/documents/{id}/duplicate
func (h *DocumentHandler) Duplicate(w http.ResponseWriter, r *http.Request) {
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

	doc, err := h.documentService.Duplicate(r.Context(), docID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.Created(w, doc)
}

// Lock handles POST /api/v1/documents/{id}/lock
func (h *DocumentHandler) Lock(w http.ResponseWriter, r *http.Request) {
	response.Error(w, apperror.BadRequest("fitur lock akan diimplementasikan di phase 14"))
}

// Sign handles POST /api/v1/documents/{id}/sign
func (h *DocumentHandler) Sign(w http.ResponseWriter, r *http.Request) {
	response.Error(w, apperror.BadRequest("fitur sign akan diimplementasikan di phase 14"))
}

// -- Block endpoints --

// AddBlock handles POST /api/v1/documents/{id}/blocks
func (h *DocumentHandler) AddBlock(w http.ResponseWriter, r *http.Request) {
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

	var req addBlockRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	input := service.AddBlockInput{
		Type:     model.BlockType(req.Type),
		Content:  req.Content,
		Position: req.Position,
		Checked:  req.Checked,
		Rows:     req.Rows,
		Columns:  req.Columns,
		Language: req.Language,
		Emoji:    req.Emoji,
		Color:    req.Color,
	}

	if req.ParentBlockID != nil {
		parentID, err := uuid.Parse(*req.ParentBlockID)
		if err != nil {
			response.Error(w, apperror.BadRequest("format parentBlockId tidak valid"))
			return
		}
		input.ParentBlockID = &parentID
	}

	block, err := h.blockService.AddBlock(r.Context(), docID, userID, input)
	if err != nil {
		handleError(w, err)
		return
	}

	response.Created(w, block)
}

// UpdateBlock handles PUT /api/v1/documents/{id}/blocks/{blockId}
func (h *DocumentHandler) UpdateBlock(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	blockID, err := uuid.Parse(chi.URLParam(r, "blockId"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format block ID tidak valid"))
		return
	}

	var req updateBlockRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	block, err := h.blockService.UpdateBlock(r.Context(), blockID, userID, model.UpdateBlockInput{
		Content:  req.Content,
		Checked:  req.Checked,
		Rows:     req.Rows,
		Columns:  req.Columns,
		Language: req.Language,
		Emoji:    req.Emoji,
		Color:    req.Color,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, block)
}

// DeleteBlock handles DELETE /api/v1/documents/{id}/blocks/{blockId}
func (h *DocumentHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	blockID, err := uuid.Parse(chi.URLParam(r, "blockId"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format block ID tidak valid"))
		return
	}

	if err := h.blockService.DeleteBlock(r.Context(), blockID, userID); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"deleted": true})
}

// ReorderBlocks handles PUT /api/v1/documents/{id}/blocks/reorder
func (h *DocumentHandler) ReorderBlocks(w http.ResponseWriter, r *http.Request) {
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

	var req reorderBlocksRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	blockIDs := make([]uuid.UUID, 0, len(req.BlockIDs))
	for _, idStr := range req.BlockIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			response.Error(w, apperror.BadRequest("format block ID tidak valid: "+idStr))
			return
		}
		blockIDs = append(blockIDs, id)
	}

	if err := h.blockService.ReorderBlocks(r.Context(), docID, userID, blockIDs); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"reordered": true})
}

// BatchBlocks handles POST /api/v1/documents/{id}/blocks/batch
func (h *DocumentHandler) BatchBlocks(w http.ResponseWriter, r *http.Request) {
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

	var req batchBlockRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	if err := h.blockService.BatchUpdate(r.Context(), docID, userID, req.Operations); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"updated": true})
}

// -- Collaborator endpoints --

// AddCollaborator handles POST /api/v1/documents/{id}/collaborators
func (h *DocumentHandler) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	ownerID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	var req addCollaboratorRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(w, apperror.BadRequest("format userId tidak valid"))
		return
	}

	role := model.CollaboratorRole(req.Role)
	if role != model.CollaboratorRoleEditor && role != model.CollaboratorRoleViewer {
		response.Error(w, apperror.BadRequest("role harus 'editor' atau 'viewer'"))
		return
	}

	if err := h.documentService.AddCollaborator(r.Context(), docID, ownerID, userID, role); err != nil {
		handleError(w, err)
		return
	}

	response.Created(w, map[string]bool{"added": true})
}

// RemoveCollaborator handles DELETE /api/v1/documents/{id}/collaborators/{userID}
func (h *DocumentHandler) RemoveCollaborator(w http.ResponseWriter, r *http.Request) {
	ownerID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format user ID tidak valid"))
		return
	}

	if err := h.documentService.RemoveCollaborator(r.Context(), docID, ownerID, userID); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"removed": true})
}

// UpdateCollaboratorRole handles PUT /api/v1/documents/{id}/collaborators/{userID}
func (h *DocumentHandler) UpdateCollaboratorRole(w http.ResponseWriter, r *http.Request) {
	ownerID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userID"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format user ID tidak valid"))
		return
	}

	var req updateCollaboratorRoleRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	role := model.CollaboratorRole(req.Role)
	if role != model.CollaboratorRoleEditor && role != model.CollaboratorRoleViewer {
		response.Error(w, apperror.BadRequest("role harus 'editor' atau 'viewer'"))
		return
	}

	if err := h.documentService.UpdateCollaboratorRole(r.Context(), docID, ownerID, userID, role); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"updated": true})
}

// -- Tag endpoints --

// AddTag handles POST /api/v1/documents/{id}/tags
func (h *DocumentHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	var req addTagRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("body request tidak valid"))
		return
	}

	if err := h.documentService.AddTag(r.Context(), docID, req.Tag); err != nil {
		handleError(w, err)
		return
	}

	response.Created(w, map[string]bool{"added": true})
}

// RemoveTag handles DELETE /api/v1/documents/{id}/tags/{tag}
func (h *DocumentHandler) RemoveTag(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	tag := chi.URLParam(r, "tag")
	if tag == "" {
		response.Error(w, apperror.BadRequest("tag tidak boleh kosong"))
		return
	}

	if err := h.documentService.RemoveTag(r.Context(), docID, tag); err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, map[string]bool{"removed": true})
}

// -- History endpoint --

// GetHistory handles GET /api/v1/documents/{id}/history
func (h *DocumentHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	docID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("format document ID tidak valid"))
		return
	}

	history, err := h.documentService.GetHistory(r.Context(), docID)
	if err != nil {
		handleError(w, err)
		return
	}

	response.OK(w, history)
}

// -- Template endpoints --

// ListTemplates handles GET /api/v1/templates
func (h *DocumentHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates := h.templateService.GetTemplates()
	response.OK(w, templates)
}

// -- Helper --

func handleError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		response.Error(w, appErr)
		return
	}
	response.Error(w, apperror.Internal(fmt.Errorf("kesalahan internal server")))
}
