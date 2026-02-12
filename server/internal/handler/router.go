package handler

import (
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/otoritech/chatat/internal/config"
	mw "github.com/otoritech/chatat/internal/middleware"
)

// NewRouter creates and configures the Chi router with all middleware and routes.
func NewRouter(cfg *config.Config, deps *Dependencies) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(mw.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(mw.RateLimit(100.0/60.0, 100)) // 100 req/min

	// Health check
	r.Get("/health", HealthCheck)

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (auth)
		r.Group(func(r chi.Router) {
			r.Use(mw.RateLimit(5.0/60.0, 5)) // 5 req/min for auth
			r.Post("/auth/otp/send", deps.AuthHandler.SendOTP)
			r.Post("/auth/otp/verify", deps.AuthHandler.VerifyOTP)
			r.Post("/auth/reverse-otp/init", deps.AuthHandler.InitReverseOTP)
			r.Post("/auth/reverse-otp/check", deps.AuthHandler.CheckReverseOTP)
			r.Post("/auth/refresh", deps.AuthHandler.RefreshToken)
		})

		// Webhook routes (server-to-server, no user auth)
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/whatsapp", deps.WebhookHandler.HandleWhatsApp)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(mw.Auth(cfg.JWTSecret))

			// Auth (requires token)
			r.Post("/auth/logout", deps.AuthHandler.Logout)

			r.Route("/users", func(r chi.Router) {
				r.Get("/me", deps.UserHandler.GetMe)
				r.Put("/me", deps.UserHandler.UpdateMe)
			})

			r.Route("/contacts", func(r chi.Router) {
				r.Post("/sync", deps.ContactHandler.Sync)
				r.Get("/", deps.ContactHandler.List)
			})

			r.Route("/chats", func(r chi.Router) {
				r.Get("/", deps.ChatHandler.List)
				r.Post("/", deps.ChatHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.ChatHandler.GetByID)
					r.Put("/", deps.ChatHandler.Update)
					r.Delete("/", deps.ChatHandler.Delete)
					r.Post("/messages", deps.ChatHandler.SendMessage)
					r.Get("/messages", deps.ChatHandler.ListMessages)
					r.Post("/members", deps.ChatHandler.AddMember)
					r.Delete("/members/{memberID}", deps.ChatHandler.RemoveMember)
				})
			})

			r.Route("/topics", func(r chi.Router) {
				r.Post("/", deps.TopicHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.TopicHandler.GetByID)
					r.Put("/", deps.TopicHandler.Update)
					r.Delete("/", deps.TopicHandler.Delete)
					r.Post("/messages", deps.TopicHandler.SendMessage)
					r.Get("/messages", deps.TopicHandler.ListMessages)
				})
			})

			r.Route("/documents", func(r chi.Router) {
				r.Get("/", deps.DocumentHandler.List)
				r.Post("/", deps.DocumentHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.DocumentHandler.GetByID)
					r.Put("/", deps.DocumentHandler.Update)
					r.Delete("/", deps.DocumentHandler.Delete)
					r.Post("/lock", deps.DocumentHandler.Lock)
					r.Post("/sign", deps.DocumentHandler.Sign)
					r.Post("/collaborators", deps.DocumentHandler.AddCollaborator)
					r.Delete("/collaborators/{userID}", deps.DocumentHandler.RemoveCollaborator)
				})
			})

			r.Route("/entities", func(r chi.Router) {
				r.Get("/", deps.EntityHandler.List)
				r.Post("/", deps.EntityHandler.Create)
				r.Get("/search", deps.EntityHandler.Search)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.EntityHandler.GetByID)
					r.Put("/", deps.EntityHandler.Update)
					r.Delete("/", deps.EntityHandler.Delete)
				})
			})
		})
	})

	// WebSocket
	r.Get("/ws", deps.WSHandler.HandleConnection)

	return r
}
