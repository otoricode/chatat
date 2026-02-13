package handler

import (
	"context"

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
	StorageService      service.StorageService
	ImageService        service.ImageService
	MediaService        service.MediaService
	DocumentService     service.DocumentService
	BlockService        service.BlockService
	TemplateService     service.TemplateService
	NotificationService service.NotificationService
	SearchService       service.SearchService

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
	MediaRepo       repository.MediaRepository
	DeviceTokenRepo repository.DeviceTokenRepository
	SearchRepo      repository.SearchRepository

	// Handlers
	AuthHandler         *AuthHandler
	WebhookHandler      *WebhookHandler
	UserHandler         *UserHandler
	ContactHandler      *ContactHandler
	ChatHandler         *ChatHandler
	TopicHandler        *TopicHandler
	MediaHandler        *MediaHandler
	DocumentHandler     *DocumentHandler
	EntityHandler       *EntityHandler
	NotificationHandler *NotificationHandler
	SearchHandler       *SearchHandler
	WSHandler           *WSHandler
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
	mediaRepo := repository.NewMediaRepository(db)
	deviceTokenRepo := repository.NewDeviceTokenRepository(db)
	searchRepo := repository.NewSearchRepository(db)

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

	// Push notification service (created early so other services can use it)
	var pushSender service.PushSender
	if cfg.FCMCredentialsFile != "" {
		fcmSender, fcmErr := service.NewFCMPushSender(context.Background(), cfg.FCMCredentialsFile)
		if fcmErr != nil {
			panic("failed to create FCM sender: " + fcmErr.Error())
		}
		pushSender = fcmSender
	} else {
		pushSender = service.NewLogPushSender()
	}
	notifSvc := service.NewNotificationService(deviceTokenRepo, chatRepo, pushSender)

	contactService := service.NewContactService(userRepo, contactRepo, hub)
	chatService := service.NewChatService(chatRepo, messageRepo, messageStatRepo, userRepo, hub)
	messageService := service.NewMessageService(messageRepo, messageStatRepo, chatRepo, userRepo, hub, notifSvc)
	groupService := service.NewGroupService(chatRepo, messageRepo, messageStatRepo, userRepo, hub, notifSvc)
	topicService := service.NewTopicService(topicRepo, topicMsgRepo, chatRepo, userRepo, hub)
	topicMsgService := service.NewTopicMessageService(topicMsgRepo, topicRepo, hub)
	storageSvc, err := service.NewStorageService(cfg)
	if err != nil {
		panic("failed to create storage service: " + err.Error())
	}
	imageSvc := service.NewImageService()
	mediaSvc := service.NewMediaService(mediaRepo, storageSvc, imageSvc)
	templateSvc := service.NewTemplateService()
	documentSvc := service.NewDocumentService(documentRepo, blockRepo, docHistoryRepo, userRepo, templateSvc, notifSvc)
	blockSvc := service.NewBlockService(blockRepo, documentRepo, docHistoryRepo)

	// Status notifier: broadcasts online/offline events to contacts
	_ = service.NewStatusNotifier(hub, contactRepo, userRepo, redisClient)

	// Auth handler
	authHandler := NewAuthHandler(otpService, reverseOTPService, tokenService, sessionService, userRepo)
	webhookHandler := NewWebhookHandler(reverseOTPService, cfg.WAWebhookSecret)
	userHandler := NewUserHandler(userService)
	contactHandler := NewContactHandler(contactService)
	chatHandler := NewChatHandler(chatService, messageService, groupService)
	topicHandler := NewTopicHandler(topicService, topicMsgService)
	mediaHandler := NewMediaHandler(mediaSvc)
	documentHandler := NewDocumentHandler(documentSvc, blockSvc, templateSvc)
	entitySvc := service.NewEntityService(entityRepo, userRepo, documentRepo)
	entityHandler := NewEntityHandler(entitySvc)
	notifHandler := NewNotificationHandler(notifSvc)
	searchSvc := service.NewSearchService(searchRepo, chatRepo)
	searchHandler := NewSearchHandler(searchSvc)

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
		StorageService:      storageSvc,
		ImageService:        imageSvc,
		MediaService:        mediaSvc,
		DocumentService:     documentSvc,
		BlockService:        blockSvc,
		TemplateService:     templateSvc,
		NotificationService: notifSvc,
		SearchService:       searchSvc,

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
		MediaRepo:       mediaRepo,
		DeviceTokenRepo: deviceTokenRepo,
		SearchRepo:      searchRepo,

		AuthHandler:         authHandler,
		WebhookHandler:      webhookHandler,
		UserHandler:         userHandler,
		ContactHandler:      contactHandler,
		ChatHandler:         chatHandler,
		TopicHandler:        topicHandler,
		MediaHandler:        mediaHandler,
		DocumentHandler:     documentHandler,
		EntityHandler:       entityHandler,
		NotificationHandler: notifHandler,
		SearchHandler:       searchHandler,
		WSHandler:           NewWSHandler(hub, cfg.JWTSecret, chatRepo, topicRepo, messageStatRepo, redisClient),
	}

	return deps
}
