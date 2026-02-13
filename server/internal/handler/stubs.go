package handler

import (
	"net/http"

	"github.com/otoritech/chatat/pkg/response"
)

type notImplementedResponse struct {
	Message string `json:"message"`
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	response.OK(w, notImplementedResponse{Message: "not implemented yet"})
}

// ChatStubHandler is a placeholder for chat endpoints.
type ChatStubHandler struct{}

func (h *ChatStubHandler) List(w http.ResponseWriter, r *http.Request)         { notImplemented(w, r) }
func (h *ChatStubHandler) Create(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *ChatStubHandler) GetByID(w http.ResponseWriter, r *http.Request)      { notImplemented(w, r) }
func (h *ChatStubHandler) Update(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *ChatStubHandler) Delete(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *ChatStubHandler) SendMessage(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *ChatStubHandler) ListMessages(w http.ResponseWriter, r *http.Request) { notImplemented(w, r) }
func (h *ChatStubHandler) AddMember(w http.ResponseWriter, r *http.Request)    { notImplemented(w, r) }
func (h *ChatStubHandler) RemoveMember(w http.ResponseWriter, r *http.Request) { notImplemented(w, r) }

// TopicStubHandler is a placeholder for topic endpoints.
type TopicStubHandler struct{}

func (h *TopicStubHandler) Create(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *TopicStubHandler) GetByID(w http.ResponseWriter, r *http.Request)      { notImplemented(w, r) }
func (h *TopicStubHandler) Update(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *TopicStubHandler) Delete(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *TopicStubHandler) SendMessage(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *TopicStubHandler) ListMessages(w http.ResponseWriter, r *http.Request) { notImplemented(w, r) }

// EntityStubHandler is a placeholder for entity endpoints.
type EntityStubHandler struct{}

func (h *EntityStubHandler) List(w http.ResponseWriter, r *http.Request)    { notImplemented(w, r) }
func (h *EntityStubHandler) Create(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *EntityStubHandler) Search(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *EntityStubHandler) GetByID(w http.ResponseWriter, r *http.Request) { notImplemented(w, r) }
func (h *EntityStubHandler) Update(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *EntityStubHandler) Delete(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
