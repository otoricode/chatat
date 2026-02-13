package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/pkg/apperror"
	"github.com/otoritech/chatat/pkg/response"
)

// SearchHandler handles search HTTP endpoints.
type SearchHandler struct {
	service service.SearchService
}

// NewSearchHandler creates a new search handler.
func NewSearchHandler(svc service.SearchService) *SearchHandler {
	return &SearchHandler{service: svc}
}

// SearchAll handles GET /search?q=query&limit=3
func (h *SearchHandler) SearchAll(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		response.Error(w, apperror.BadRequest("kata kunci pencarian minimal 2 karakter"))
		return
	}

	limit := 3
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}

	results, err := h.service.SearchAll(r.Context(), userID, query, limit)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, results)
}

// SearchMessages handles GET /search/messages?q=query&offset=0&limit=20
func (h *SearchHandler) SearchMessages(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		response.Error(w, apperror.BadRequest("kata kunci pencarian minimal 2 karakter"))
		return
	}

	opts := parseSearchOpts(r)
	results, err := h.service.SearchMessages(r.Context(), userID, query, opts)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, results)
}

// SearchDocuments handles GET /search/documents?q=query&offset=0&limit=20
func (h *SearchHandler) SearchDocuments(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		response.Error(w, apperror.BadRequest("kata kunci pencarian minimal 2 karakter"))
		return
	}

	opts := parseSearchOpts(r)
	results, err := h.service.SearchDocuments(r.Context(), userID, query, opts)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, results)
}

// SearchContacts handles GET /search/contacts?q=query
func (h *SearchHandler) SearchContacts(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		response.Error(w, apperror.BadRequest("kata kunci pencarian minimal 2 karakter"))
		return
	}

	results, err := h.service.SearchContacts(r.Context(), userID, query)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, results)
}

// SearchEntities handles GET /search/entities?q=query
func (h *SearchHandler) SearchEntities(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		response.Error(w, apperror.BadRequest("kata kunci pencarian minimal 2 karakter"))
		return
	}

	results, err := h.service.SearchEntities(r.Context(), userID, query)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, results)
}

// SearchInChat handles GET /chats/{id}/search?q=query&offset=0&limit=20
func (h *SearchHandler) SearchInChat(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserID(r)
	if err != nil {
		response.Error(w, apperror.Unauthorized("autentikasi diperlukan"))
		return
	}

	chatID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperror.BadRequest("ID chat tidak valid"))
		return
	}

	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		response.Error(w, apperror.BadRequest("kata kunci pencarian minimal 2 karakter"))
		return
	}

	opts := parseSearchOpts(r)
	results, err := h.service.SearchInChat(r.Context(), chatID, userID, query, opts)
	if err != nil {
		response.Error(w, apperror.Internal(err))
		return
	}

	response.OK(w, results)
}

func parseSearchOpts(r *http.Request) service.SearchOpts {
	opts := service.SearchOpts{
		Offset: 0,
		Limit:  20,
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			opts.Offset = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			opts.Limit = v
		}
	}
	return opts
}
