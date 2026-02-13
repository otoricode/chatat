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
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Accept-Language"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(mw.RateLimit(100.0/60.0, 100)) // 100 req/min
	r.Use(mw.Language())

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
				r.Post("/me/setup", deps.UserHandler.SetupProfile)
				r.Delete("/me", deps.UserHandler.DeleteAccount)
			})

			r.Route("/contacts", func(r chi.Router) {
				r.Post("/sync", deps.ContactHandler.Sync)
				r.Get("/", deps.ContactHandler.List)
				r.Get("/search", deps.ContactHandler.Search)
				r.Get("/{userId}", deps.ContactHandler.GetProfile)
			})

			r.Route("/chats", func(r chi.Router) {
				r.Get("/", deps.ChatHandler.List)
				r.Post("/", deps.ChatHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.ChatHandler.GetByID)
					r.Put("/", deps.ChatHandler.Update)
					r.Delete("/", deps.ChatHandler.Delete)
					r.Put("/pin", deps.ChatHandler.PinChat)
					r.Delete("/pin", deps.ChatHandler.UnpinChat)
					r.Post("/read", deps.ChatHandler.MarkAsRead)
					r.Get("/info", deps.ChatHandler.GetGroupInfo)
					r.Post("/leave", deps.ChatHandler.LeaveGroup)
					r.Post("/messages", deps.ChatHandler.SendMessage)
					r.Get("/messages", deps.ChatHandler.ListMessages)
					r.Get("/messages/search", deps.ChatHandler.SearchMessages)
					r.Delete("/messages/{messageId}", deps.ChatHandler.DeleteMessage)
					r.Post("/messages/{messageId}/forward", deps.ChatHandler.ForwardMessage)
					r.Post("/members", deps.ChatHandler.AddMember)
					r.Delete("/members/{memberID}", deps.ChatHandler.RemoveMember)
					r.Put("/members/{memberID}/admin", deps.ChatHandler.PromoteToAdmin)
					r.Get("/topics", deps.TopicHandler.ListByChat)
					r.Get("/documents", deps.DocumentHandler.ListByChat)
					r.Get("/search", deps.SearchHandler.SearchInChat)
				})
			})

			r.Route("/topics", func(r chi.Router) {
				r.Get("/", deps.TopicHandler.ListByUser)
				r.Post("/", deps.TopicHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.TopicHandler.GetByID)
					r.Put("/", deps.TopicHandler.Update)
					r.Delete("/", deps.TopicHandler.Delete)
					r.Post("/members", deps.TopicHandler.AddMember)
					r.Delete("/members/{userId}", deps.TopicHandler.RemoveMember)
					r.Post("/messages", deps.TopicHandler.SendMessage)
					r.Get("/messages", deps.TopicHandler.ListMessages)
					r.Delete("/messages/{messageId}", deps.TopicHandler.DeleteMessage)
					r.Get("/documents", deps.DocumentHandler.ListByTopic)
				})
			})

			r.Route("/media", func(r chi.Router) {
				r.Post("/upload", deps.MediaHandler.Upload)
				r.Get("/{id}", deps.MediaHandler.GetByID)
				r.Get("/{id}/download", deps.MediaHandler.Download)
				r.Delete("/{id}", deps.MediaHandler.Delete)
			})

			r.Route("/documents", func(r chi.Router) {
				r.Get("/", deps.DocumentHandler.List)
				r.Post("/", deps.DocumentHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.DocumentHandler.GetByID)
					r.Put("/", deps.DocumentHandler.Update)
					r.Delete("/", deps.DocumentHandler.Delete)
					r.Post("/duplicate", deps.DocumentHandler.Duplicate)
					r.Post("/lock", deps.DocumentHandler.Lock)
					r.Post("/unlock", deps.DocumentHandler.Unlock)
					r.Post("/sign", deps.DocumentHandler.Sign)

					// Signer endpoints
					r.Get("/signers", deps.DocumentHandler.ListSigners)
					r.Post("/signers", deps.DocumentHandler.AddSigner)
					r.Delete("/signers/{userID}", deps.DocumentHandler.RemoveSigner)

					// Block endpoints
					r.Post("/blocks", deps.DocumentHandler.AddBlock)
					r.Put("/blocks/reorder", deps.DocumentHandler.ReorderBlocks)
					r.Post("/blocks/batch", deps.DocumentHandler.BatchBlocks)
					r.Put("/blocks/{blockId}", deps.DocumentHandler.UpdateBlock)
					r.Delete("/blocks/{blockId}", deps.DocumentHandler.DeleteBlock)

					// Collaborator endpoints
					r.Post("/collaborators", deps.DocumentHandler.AddCollaborator)
					r.Put("/collaborators/{userID}", deps.DocumentHandler.UpdateCollaboratorRole)
					r.Delete("/collaborators/{userID}", deps.DocumentHandler.RemoveCollaborator)

					// Tag endpoints
					r.Post("/tags", deps.DocumentHandler.AddTag)
					r.Delete("/tags/{tag}", deps.DocumentHandler.RemoveTag)

					// History endpoint
					r.Get("/history", deps.DocumentHandler.GetHistory)

					// Entity linking endpoints
					r.Get("/entities", deps.EntityHandler.GetDocumentEntities)
					r.Post("/entities", deps.EntityHandler.LinkToDocument)
					r.Delete("/entities/{entityId}", deps.EntityHandler.UnlinkFromDocument)
				})
			})

			r.Get("/templates", deps.DocumentHandler.ListTemplates)

			r.Route("/entities", func(r chi.Router) {
				r.Get("/", deps.EntityHandler.List)
				r.Post("/", deps.EntityHandler.Create)
				r.Get("/search", deps.EntityHandler.Search)
				r.Get("/types", deps.EntityHandler.ListTypes)
				r.Post("/from-contact", deps.EntityHandler.CreateFromContact)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", deps.EntityHandler.GetByID)
					r.Put("/", deps.EntityHandler.Update)
					r.Delete("/", deps.EntityHandler.Delete)
					r.Get("/documents", deps.EntityHandler.ListDocuments)
				})
			})

			r.Route("/notifications", func(r chi.Router) {
				r.Post("/devices", deps.NotificationHandler.RegisterDevice)
				r.Delete("/devices", deps.NotificationHandler.UnregisterDevice)
			})

			r.Route("/search", func(r chi.Router) {
				r.Get("/", deps.SearchHandler.SearchAll)
				r.Get("/messages", deps.SearchHandler.SearchMessages)
				r.Get("/documents", deps.SearchHandler.SearchDocuments)
				r.Get("/contacts", deps.SearchHandler.SearchContacts)
				r.Get("/entities", deps.SearchHandler.SearchEntities)
			})

			r.Route("/backup", func(r chi.Router) {
				r.Get("/export", deps.BackupHandler.Export)
				r.Post("/import", deps.BackupHandler.Import)
				r.Post("/log", deps.BackupHandler.LogBackup)
				r.Get("/history", deps.BackupHandler.GetHistory)
				r.Get("/latest", deps.BackupHandler.GetLatest)
			})
		})
	})

	// WebSocket
	r.Get("/ws", deps.WSHandler.HandleConnection)

	return r
}
