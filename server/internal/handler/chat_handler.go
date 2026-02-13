package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// ChatHandler handles chat endpoints.
type ChatHandler struct {
	chatService    service.ChatService
	messageService service.MessageService
	groupService   service.GroupService
}

// NewChatHandler creates a new chat handler.
func NewChatHandler(chatService service.ChatService, messageService service.MessageService, groupService service.GroupService) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		messageService: messageService,
		groupService:   groupService,
	}
}

type createChatRequest struct {
	Type        string   `json:"type"`
	ContactID   string   `json:"contactId"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Description string   `json:"description"`
	MemberIDs   []string `json:"memberIds"`
}

// List handles GET /api/v1/chats
func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chats, err := h.chatService.ListChats(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, chats)
}

// Create handles POST /api/v1/chats
func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var req createChatRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	// Group chat creation
	if req.Type == "group" {
		memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
		for _, idStr := range req.MemberIDs {
			id, err := parseUUID(idStr)
			if err != nil {
				response.Error(w, apperror.BadRequest("invalid member id: "+idStr))
				return
			}
			memberIDs = append(memberIDs, id)
		}

		chat, err := h.groupService.CreateGroup(r.Context(), userID, service.CreateGroupInput{
			Name:        req.Name,
			Icon:        req.Icon,
			Description: req.Description,
			MemberIDs:   memberIDs,
		})
		if err != nil {
			handleServiceError(w, err)
			return
		}
		response.Created(w, chat)
		return
	}

	// Personal chat creation
	if req.ContactID == "" {
		response.Error(w, apperror.BadRequest("contactId is required for personal chat"))
		return
	}

	contactID, err := parseUUID(req.ContactID)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid contactId format"))
		return
	}

	chat, err := h.chatService.GetOrCreatePersonalChat(r.Context(), userID, contactID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Created(w, chat)
}

// GetByID handles GET /api/v1/chats/{id}
func (h *ChatHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	detail, err := h.chatService.GetChat(r.Context(), chatID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, detail)
}

// Update handles PUT /api/v1/chats/{id}
func (h *ChatHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	var req service.UpdateGroupInput
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	chat, err := h.groupService.UpdateGroup(r.Context(), chatID, userID, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, chat)
}

// Delete handles DELETE /api/v1/chats/{id}
func (h *ChatHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	if err := h.groupService.DeleteGroup(r.Context(), chatID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// PinChat handles PUT /api/v1/chats/{id}/pin
func (h *ChatHandler) PinChat(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	if err := h.chatService.PinChat(r.Context(), chatID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// UnpinChat handles DELETE /api/v1/chats/{id}/pin
func (h *ChatHandler) UnpinChat(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	if err := h.chatService.UnpinChat(r.Context(), chatID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// SendMessage handles POST /api/v1/chats/{id}/messages
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	var req service.SendMessageInput
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	req.ChatID = chatID
	req.SenderID = userID

	msg, err := h.messageService.SendMessage(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Created(w, msg)
}

// ListMessages handles GET /api/v1/chats/{id}/messages
func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	// Verify membership
	isMember, err := h.chatService.IsMember(r.Context(), chatID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if !isMember {
		response.Error(w, apperror.Forbidden("you are not a member of this chat"))
		return
	}

	cursor, limit := ParsePagination(r)

	page, err := h.messageService.GetMessages(r.Context(), chatID, cursor, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Paginated(w, page.Messages, response.PaginationMeta{
		Cursor:  page.Cursor,
		HasMore: page.HasMore,
	})
}

// DeleteMessage handles DELETE /api/v1/chats/{id}/messages/{messageId}
func (h *ChatHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	messageID, err := GetPathUUID(r, "messageId")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid message id"))
		return
	}

	forAll := r.URL.Query().Get("forAll") == "true"

	if err := h.messageService.DeleteMessage(r.Context(), messageID, userID, forAll); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// ForwardMessage handles POST /api/v1/chats/{id}/messages/{messageId}/forward
func (h *ChatHandler) ForwardMessage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	messageID, err := GetPathUUID(r, "messageId")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid message id"))
		return
	}

	var req struct {
		TargetChatID string `json:"targetChatId"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	targetChatID, err := parseUUID(req.TargetChatID)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid targetChatId format"))
		return
	}

	msg, err := h.messageService.ForwardMessage(r.Context(), messageID, userID, targetChatID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Created(w, msg)
}

// SearchMessages handles GET /api/v1/chats/{id}/messages/search
func (h *ChatHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	// Verify membership
	isMember, err := h.chatService.IsMember(r.Context(), chatID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	if !isMember {
		response.Error(w, apperror.Forbidden("you are not a member of this chat"))
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		response.Error(w, apperror.BadRequest("search query parameter 'q' is required"))
		return
	}

	messages, err := h.messageService.SearchMessages(r.Context(), chatID, query)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, messages)
}

// AddMember handles POST /api/v1/chats/{id}/members
func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	var req struct {
		UserID string `json:"userId"`
	}
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	memberID, err := parseUUID(req.UserID)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid userId format"))
		return
	}

	if err := h.groupService.AddMember(r.Context(), chatID, memberID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// RemoveMember handles DELETE /api/v1/chats/{id}/members/{memberID}
func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	memberID, err := GetPathUUID(r, "memberID")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid member id"))
		return
	}

	if err := h.groupService.RemoveMember(r.Context(), chatID, memberID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// PromoteToAdmin handles PUT /api/v1/chats/{id}/members/{memberID}/admin
func (h *ChatHandler) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	memberID, err := GetPathUUID(r, "memberID")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid member id"))
		return
	}

	if err := h.groupService.PromoteToAdmin(r.Context(), chatID, memberID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// LeaveGroup handles POST /api/v1/chats/{id}/leave
func (h *ChatHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	if err := h.groupService.LeaveGroup(r.Context(), chatID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// GetGroupInfo handles GET /api/v1/chats/{id}/info
func (h *ChatHandler) GetGroupInfo(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	info, err := h.groupService.GetGroupInfo(r.Context(), chatID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, info)
}

// MarkAsRead handles POST /api/v1/chats/{id}/read
func (h *ChatHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	chatID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid chat id"))
		return
	}

	if err := h.messageService.MarkChatAsRead(r.Context(), chatID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.NoContent(w)
}

// parseUUID parses a string into uuid.UUID.
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// handleServiceError maps service errors to HTTP responses.
func handleServiceError(w http.ResponseWriter, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		response.Error(w, appErr)
		return
	}
	response.Error(w, apperror.Internal(err))
}
