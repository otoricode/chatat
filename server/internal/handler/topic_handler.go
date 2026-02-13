package handler

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/model"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// TopicHandler handles topic endpoints.
type TopicHandler struct {
	topicService    service.TopicService
	topicMsgService service.TopicMessageService
}

// NewTopicHandler creates a new TopicHandler.
func NewTopicHandler(topicService service.TopicService, topicMsgService service.TopicMessageService) *TopicHandler {
	return &TopicHandler{
		topicService:    topicService,
		topicMsgService: topicMsgService,
	}
}

type createTopicRequest struct {
	Name        string   `json:"name"`
	Icon        string   `json:"icon"`
	Description string   `json:"description"`
	ParentID    string   `json:"parentId"`
	MemberIDs   []string `json:"memberIds"`
}

type updateTopicRequest struct {
	Name        *string `json:"name"`
	Icon        *string `json:"icon"`
	Description *string `json:"description"`
}

type sendTopicMessageRequest struct {
	Content   string  `json:"content"`
	ReplyToID *string `json:"replyToId"`
	Type      string  `json:"type"`
}

type addTopicMemberRequest struct {
	UserID string `json:"userId"`
}

// Create handles POST /api/v1/topics
func (h *TopicHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	var req createTopicRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	parentID, err := uuid.Parse(req.ParentID)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid parentId format"))
		return
	}

	memberIDs := make([]uuid.UUID, 0, len(req.MemberIDs))
	for _, idStr := range req.MemberIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			response.Error(w, apperror.BadRequest("invalid member id: "+idStr))
			return
		}
		memberIDs = append(memberIDs, id)
	}

	topic, err := h.topicService.CreateTopic(r.Context(), userID, service.CreateTopicInput{
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
		ParentID:    parentID,
		MemberIDs:   memberIDs,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Created(w, topic)
}

// GetByID handles GET /api/v1/topics/:id
func (h *TopicHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	detail, err := h.topicService.GetTopic(r.Context(), topicID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, detail)
}

// ListByChat handles GET /api/v1/chats/:id/topics
func (h *TopicHandler) ListByChat(w http.ResponseWriter, r *http.Request) {
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

	topics, err := h.topicService.ListByChat(r.Context(), chatID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, topics)
}

// ListByUser handles GET /api/v1/topics
func (h *TopicHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topics, err := h.topicService.ListByUser(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, topics)
}

// Update handles PUT /api/v1/topics/:id
func (h *TopicHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	var req updateTopicRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	topic, err := h.topicService.UpdateTopic(r.Context(), topicID, userID, service.UpdateTopicInput{
		Name:        req.Name,
		Icon:        req.Icon,
		Description: req.Description,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, topic)
}

// Delete handles DELETE /api/v1/topics/:id
func (h *TopicHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	if err := h.topicService.DeleteTopic(r.Context(), topicID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "topic deleted"})
}

// AddMember handles POST /api/v1/topics/:id/members
func (h *TopicHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	var req addTopicMemberRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	memberID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid userId format"))
		return
	}

	if err := h.topicService.AddMember(r.Context(), topicID, memberID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "member added"})
}

// RemoveMember handles DELETE /api/v1/topics/:id/members/:userId
func (h *TopicHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	memberIDStr := GetPathParam(r, "userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid userId format"))
		return
	}

	if err := h.topicService.RemoveMember(r.Context(), topicID, memberID, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "member removed"})
}

// SendMessage handles POST /api/v1/topics/:id/messages
func (h *TopicHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	var req sendTopicMessageRequest
	if err := DecodeJSON(r, &req); err != nil {
		response.Error(w, apperror.BadRequest("invalid request body"))
		return
	}

	input := service.SendTopicMessageInput{
		TopicID:  topicID,
		SenderID: userID,
		Content:  req.Content,
		Type:     model.MessageTypeText,
	}

	if req.Type != "" {
		input.Type = model.MessageType(req.Type)
	}

	if req.ReplyToID != nil && *req.ReplyToID != "" {
		id, err := uuid.Parse(*req.ReplyToID)
		if err != nil {
			response.Error(w, apperror.BadRequest("invalid replyToId format"))
			return
		}
		input.ReplyToID = &id
	}

	msg, err := h.topicMsgService.SendMessage(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Created(w, msg)
}

// ListMessages handles GET /api/v1/topics/:id/messages
func (h *TopicHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	topicID, err := GetPathUUID(r, "id")
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid topic id"))
		return
	}

	// Verify user is member of topic
	_, err = h.topicService.GetTopic(r.Context(), topicID, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	cursor, limit := ParsePagination(r)

	page, err := h.topicMsgService.GetMessages(r.Context(), topicID, cursor, limit)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.Paginated(w, page.Messages, response.PaginationMeta{
		Cursor:  page.Cursor,
		HasMore: page.HasMore,
	})
}

// DeleteMessage handles DELETE /api/v1/topics/:id/messages/:messageId
func (h *TopicHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("user not authenticated"))
		return
	}

	messageIDStr := GetPathParam(r, "messageId")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		response.Error(w, apperror.BadRequest("invalid messageId format"))
		return
	}

	forAll := r.URL.Query().Get("forAll") == "true"

	if err := h.topicMsgService.DeleteMessage(r.Context(), messageID, userID, forAll); err != nil {
		handleServiceError(w, err)
		return
	}

	response.OK(w, map[string]string{"message": "message deleted"})
}
