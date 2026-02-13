package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/handler"
)

func TestChatStubHandler(t *testing.T) {
	h := &handler.ChatStubHandler{}
	methods := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{"List", h.List},
		{"Create", h.Create},
		{"GetByID", h.GetByID},
		{"Update", h.Update},
		{"Delete", h.Delete},
		{"SendMessage", h.SendMessage},
		{"ListMessages", h.ListMessages},
		{"AddMember", h.AddMember},
		{"RemoveMember", h.RemoveMember},
	}
	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/stub", nil)
			m.fn(w, r)
			assert.Equal(t, http.StatusNotImplemented, w.Code)
		})
	}
}

func TestTopicStubHandler(t *testing.T) {
	h := &handler.TopicStubHandler{}
	methods := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{"Create", h.Create},
		{"GetByID", h.GetByID},
		{"Update", h.Update},
		{"Delete", h.Delete},
		{"SendMessage", h.SendMessage},
		{"ListMessages", h.ListMessages},
	}
	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/stub", nil)
			m.fn(w, r)
			assert.Equal(t, http.StatusNotImplemented, w.Code)
		})
	}
}

func TestEntityStubHandler(t *testing.T) {
	h := &handler.EntityStubHandler{}
	methods := []struct {
		name string
		fn   http.HandlerFunc
	}{
		{"List", h.List},
		{"Create", h.Create},
		{"Search", h.Search},
		{"GetByID", h.GetByID},
		{"Update", h.Update},
		{"Delete", h.Delete},
	}
	for _, m := range methods {
		t.Run(m.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/stub", nil)
			m.fn(w, r)
			assert.Equal(t, http.StatusNotImplemented, w.Code)
		})
	}
}
