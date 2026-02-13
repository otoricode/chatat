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
	OTPService          service.OTPService
	ReverseOTP          service.ReverseOTPService
	TokenService        service.TokenService
	SessionService      service.SessionService
	UserService         service.UserService
	ContactService      service.ContactService
	ChatService         service.ChatService
	MessageService      service.MessageService
	GroupService        service.GroupService
	TopicService        service.TopicService
	TopicMessageService service.TopicMessageService

	// Repositories
	UserRepo        repository.UserRepository
	ContactRepo     repository.ContactRepository
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
	UserHandler     *UserHandler
	ContactHandler  *ContactHandler
	ChatHandler     *ChatHandler
	TopicHandler    *TopicHandler
	DocumentHandler *DocumentStubHandler
	EntityHandler   *EntityStubHandler
	WSHandler       *WSHandler
}

// NewDependencies creates and wires all application dependencies.
func NewDependencies(cfg *config.Config, db *pgxpool.Pool, redisClient *redis.Client, hub *ws.Hub) *Dependencies {
	// Repositories
	userRepo := repository.NewUserRepository(db)
	contactRepo := repository.NewContactRepository(db)
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

	var waProvider service.WhatsAppProvider
	if cfg.WABaseURL != "" && cfg.WABusinessPhone != "" {
		waProvider = service.NewGOWAProvider(cfg.WABaseURL, cfg.WABusinessPhone)
	} else {
		waProvider = service.NewLogWhatsAppProvider("+628001234567")
	}

	otpService := service.NewOTPService(redisClient, smsProvider, service.DefaultOTPConfig())
	reverseOTPService := service.NewReverseOTPService(redisClient, waProvider, 0)
	tokenService := service.NewTokenService(redisClient, service.DefaultTokenConfig(cfg.JWTSecret))
	sessionService := service.NewSessionService(redisClient, tokenService, 0)
	userService := service.NewUserService(userRepo)
	contactService := service.NewContactService(userRepo, contactRepo, hub)
	chatService := service.NewChatService(chatRepo, messageRepo, messageStatRepo, userRepo, hub)
	messageService := service.NewMessageService(messageRepo, messageStatRepo, chatRepo, hub)
	groupService := service.NewGroupService(chatRepo, messageRepo, messageStatRepo, userRepo, hub)
	topicService := service.NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)
	topicMsgService := service.NewTopicMessageService(topicMsgRepo, topicRepo, hub)

	// Status notifier: broadcasts online/offline events to contacts
	_ = service.NewStatusNotifier(hub, contactRepo, userRepo, redisClient)

	// Auth handler
	authHandler := NewAuthHandler(otpService, reverseOTPService, tokenService, sessionService, userRepo)
	webhookHandler := NewWebhookHandler(reverseOTPService, cfg.WAWebhookSecret)
	userHandler := NewUserHandler(userService)
	contactHandler := NewContactHandler(contactService)
	chatHandler := NewChatHandler(chatService, messageService, groupService)
	topicHandler := NewTopicHandler(topicService, topicMsgService)

	deps := &Dependencies{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		Hub:    hub,

		OTPService:          otpService,
		ReverseOTP:          reverseOTPService,
		TokenService:        tokenService,
		SessionService:      sessionService,
		UserService:         userService,
		ContactService:      contactService,
		ChatService:         chatService,
		MessageService:      messageService,
		GroupService:        groupService,
		TopicService:        topicService,
		TopicMessageService: topicMsgService,

		UserRepo:        userRepo,
		ContactRepo:     contactRepo,
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
		UserHandler:     userHandler,
		ContactHandler:  contactHandler,
		ChatHandler:     chatHandler,
		TopicHandler:    topicHandler,
		DocumentHandler: &DocumentStubHandler{},
		EntityHandler:   &EntityStubHandler{},
		WSHandler:       NewWSHandler(hub, cfg.JWTSecret, chatRepo, topicRepo, messageStatRepo, redisClient),
	}

	return deps
}
