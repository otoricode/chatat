package handler

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/otoritech/chatat/internal/config"
	"github.com/otoritech/chatat/internal/repository"
	"github.com/otoritech/chatat/internal/ws"
)

// Dependencies holds all application dependencies for dependency injection.
type Dependencies struct {
	Config *config.Config
	DB     *pgxpool.Pool
	Redis  *redis.Client
	Hub    *ws.Hub

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

	// Handlers (stubs for now, will be implemented in later phases)
	AuthHandler     *AuthStubHandler
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

	deps := &Dependencies{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
		Hub:    hub,

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

		// Stub handlers for routes not yet implemented
		AuthHandler:     &AuthStubHandler{},
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
