package middleware

import (
	"net/http"

	"github.com/otoritech/chatat/internal/i18n"
)

// Language returns middleware that extracts the language from the
// Accept-Language header and stores it in the request context.
func Language() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := i18n.ParseLang(r.Header.Get("Accept-Language"))
			ctx := i18n.WithLang(r.Context(), lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
