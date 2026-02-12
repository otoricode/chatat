package handler

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/otoritech/chatat/internal/config"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/service"
	"github.com/otoritech/chatat/internal/ws"
)

// Dependencies holds all application dependencies for dependency injection.
type Dependencies struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Hub    *ws.Hub

	// Services
	OTPService     service.OTPService
	ReverseOTP     service.ReverseOTPService
	TokenService   service.TokenService
	SessionService service.SessionService

	// Repositories
	UserRepo        repository.UserRepository
	ChatRepo        repository.ChatRepository
	MessageRepo     repository.MessageRepository
	TopicRepo       repository.TopicRepository
	DocumentRepo    repository.DocumentRepository
	BlockRepo       repository.BlockRepository
	EntityRepo      repository.EntityRepository
	MessageStatRepo repository.MessageStatusRepository
	DocHistoryRepo  repository.DocumentHistoryRepository
	TopicMsgRepo    repository.TopicMessageRepository

	// Handlers
	AuthHandler     *AuthHandler
	WebhookHandler  *WebhookHandler
	UserHandler     *UserStubHandler
	ContactHandler  *ContactStubHandler
	ChatHandler     *ChatStubHandler
	TopicHandler    *TopicStubHandler
	DocumentHandler *DocumentStubHandler
	EntityHandler   *EntityStubHandler
	WSHandler       *WSHandler
}

// NewDependencies creates and wires all application dependencies.
func NewDependencies(cfg *config.Config, db *pgxpool.Pool, redisClient *redis.Client, hub *ws.Hub) *Dependencies {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	chatRepo := repository.NewChatRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	topicRepo := repository.NewTopicRepository(db)
	documentRepo := repository.NewDocumentRepository(db)
	blockRepo := repository.NewBlockRepository(db)
	entityRepo := repository.NewEntityRepository(db)
	messageStatRepo := repository.NewMessageStatusRepository(db)
	docHistoryRepo := repository.NewDocumentHistoryRepository(db)
	topicMsgRepo := repository.NewTopicMessageRepository(db)

	// Services
	smsProvider := service.NewLogSMSProvider()
	waProvider := service.NewLogWhatsAppProvider("+628001234567")

	otpService := service.NewOTPService(redisClient, smsProvider, service.DefaultOTPConfig())
	reverseOTPService := service.NewReverseOTPService(redisClient, waProvider, 0)
	tokenService := service.NewTokenService(redisClient, service.DefaultTokenConfig(cfg.JWTSecret))
	sessionService := service.NewSessionService(redisClient, tokenService, 0)

	// Auth handler
	authHandler := NewAuthHandler(otpService, reverseOTPService, tokenService, sessionService, userRepo)
	webhookHandler := NewWebhookHandler(reverseOTPService)

	deps := &Dependencies{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		Hub:    hub,

		OTPService:     otpService,
		ReverseOTP:     reverseOTPService,
		TokenService:   tokenService,
		SessionService: sessionService,

		UserRepo:        userRepo,
		ChatRepo:        chatRepo,
		MessageRepo:     messageRepo,
		TopicRepo:       topicRepo,
		DocumentRepo:    documentRepo,
		BlockRepo:       blockRepo,
		EntityRepo:      entityRepo,
		MessageStatRepo: messageStatRepo,
		DocHistoryRepo:  docHistoryRepo,
		TopicMsgRepo:    topicMsgRepo,

		AuthHandler:     authHandler,
		WebhookHandler:  webhookHandler,
		UserHandler:     &UserStubHandler{},
		ContactHandler:  &ContactStubHandler{},
		ChatHandler:     &ChatStubHandler{},
		TopicHandler:    &TopicStubHandler{},
		DocumentHandler: &DocumentStubHandler{},
		EntityHandler:   &EntityStubHandler{},
		WSHandler:       NewWSHandler(hub, cfg.JWTSecret),
	}

	return deps
}
