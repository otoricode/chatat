package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/otoritech/chatat/internal/i18n"
	"github.com/otoritech/chatat/internal/middleware"
)

func TestLanguage_DefaultsToID(t *testing.T) {
	var gotLang i18n.Language
	handler := middleware.Language()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLang = i18n.GetLang(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, gotLang)
}

func TestLanguage_EnglishHeader(t *testing.T) {
	var gotLang i18n.Language
	handler := middleware.Language()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLang = i18n.GetLang(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Language", "en")
	handler.ServeHTTP(w, r)

	assert.Equal(t, i18n.Language("en"), gotLang)
}

func TestLanguage_IndonesianHeader(t *testing.T) {
	var gotLang i18n.Language
	handler := middleware.Language()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLang = i18n.GetLang(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Language", "id")
	handler.ServeHTTP(w, r)

	assert.Equal(t, i18n.Language("id"), gotLang)
}
