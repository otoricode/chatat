# Phase 03: API & WebSocket Foundation

> Setup HTTP router, middleware stack, error handling, dan WebSocket hub.
> Phase ini menghasilkan fondasi API dan real-time communication.

**Estimasi:** 3 hari
**Dependency:** Phase 01 (Project Setup)
**Output:** REST API framework dan WebSocket server siap digunakan.

---

## Task 3.1: HTTP Router & Middleware

**Input:** Go server skeleton dari Phase 01
**Output:** Chi router dengan middleware stack lengkap

### Steps:
1. Buat `internal/handler/router.go`:
   ```go
   func NewRouter(cfg *config.Config, deps *Dependencies) *chi.Mux {
       r := chi.NewRouter()

       // Global middleware
       r.Use(middleware.RequestID)
       r.Use(middleware.RealIP)
       r.Use(middleware.Logger)       // custom zerolog
       r.Use(middleware.Recoverer)
       r.Use(middleware.Timeout(30 * time.Second))
       r.Use(corsMiddleware(cfg))
       r.Use(rateLimitMiddleware())

       // Health check
       r.Get("/health", healthHandler)

       // API v1
       r.Route("/api/v1", func(r chi.Router) {
           // Public routes
           r.Group(func(r chi.Router) {
               r.Post("/auth/otp/send", deps.AuthHandler.SendOTP)
               r.Post("/auth/otp/verify", deps.AuthHandler.VerifyOTP)
               r.Post("/auth/reverse-otp/init", deps.AuthHandler.InitReverseOTP)
               r.Post("/auth/reverse-otp/check", deps.AuthHandler.CheckReverseOTP)
           })

           // Protected routes
           r.Group(func(r chi.Router) {
               r.Use(authMiddleware(cfg.JWTSecret))

               r.Route("/users", func(r chi.Router) { /* ... */ })
               r.Route("/contacts", func(r chi.Router) { /* ... */ })
               r.Route("/chats", func(r chi.Router) { /* ... */ })
               r.Route("/topics", func(r chi.Router) { /* ... */ })
               r.Route("/documents", func(r chi.Router) { /* ... */ })
               r.Route("/entities", func(r chi.Router) { /* ... */ })
               r.Route("/media", func(r chi.Router) { /* ... */ })
           })
       })

       // WebSocket
       r.Get("/ws", deps.WSHandler.HandleConnection)

       return r
   }
   ```
2. Buat `internal/middleware/auth.go`:
   - Extract JWT dari `Authorization: Bearer <token>` header
   - Validate token signature dan expiry
   - Set user ID di context: `context.WithValue(ctx, userIDKey, claims.UserID)`
   - Return 401 jika invalid
3. Buat `internal/middleware/cors.go`:
   - Allow origins dari config
   - Allow methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
   - Allow headers: Authorization, Content-Type
4. Buat `internal/middleware/rate_limit.go`:
   - Token bucket per IP
   - Default: 100 requests/menit
   - Auth endpoints: 5 requests/menit
   - Menggunakan Redis counter
5. Buat `internal/middleware/logger.go`:
   - Log request method, path, status, duration
   - Menggunakan zerolog
   - Skip health check dari log

### Acceptance Criteria:
- [x] Router terdefinisi dengan semua route groups
- [x] Auth middleware: block unauthorized, pass authorized
- [x] CORS middleware: proper headers set
- [x] Rate limit: enforce limit, return 429
- [x] Logger: request/response logging
- [x] Recovery: panic tidak crash server

### Testing:
- [x] Unit test: auth middleware (valid token, invalid token, expired token)
- [x] Unit test: rate limit middleware
- [x] Integration test: full request cycle

---

## Task 3.2: Error Handling & Response Pattern

**Input:** Task 3.1 selesai
**Output:** Standardized error handling dan response format

### Steps:
1. Update `internal/errors/errors.go`:
   ```go
   type AppError struct {
       Code    string `json:"code"`
       Message string `json:"message"`
       Status  int    `json:"-"`
       Err     error  `json:"-"`
   }

   // Predefined errors
   var (
       ErrNotFound       = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
       ErrUnauthorized   = &AppError{Code: "UNAUTHORIZED", Message: "Authentication required", Status: 401}
       ErrForbidden      = &AppError{Code: "FORBIDDEN", Message: "Access denied", Status: 403}
       ErrBadRequest     = &AppError{Code: "BAD_REQUEST", Message: "Invalid request", Status: 400}
       ErrConflict       = &AppError{Code: "CONFLICT", Message: "Resource already exists", Status: 409}
       ErrRateLimited    = &AppError{Code: "RATE_LIMITED", Message: "Too many requests", Status: 429}
       ErrInternal       = &AppError{Code: "INTERNAL", Message: "Internal server error", Status: 500}
       ErrDocLocked      = &AppError{Code: "DOC_LOCKED", Message: "Document is locked", Status: 423}
       ErrInvalidOTP     = &AppError{Code: "INVALID_OTP", Message: "OTP is invalid or expired", Status: 400}
   )

   func (e *AppError) Error() string { return e.Message }
   func (e *AppError) WithMessage(msg string) *AppError { ... }
   func (e *AppError) Wrap(err error) *AppError { ... }
   ```
2. Update `pkg/response/response.go`:
   ```go
   type SuccessResponse struct {
       Success bool        `json:"success"`
       Data    interface{} `json:"data,omitempty"`
   }

   type ErrorResponse struct {
       Success bool   `json:"success"`
       Error   struct {
           Code    string `json:"code"`
           Message string `json:"message"`
       } `json:"error"`
   }

   type PaginatedResponse struct {
       Success bool        `json:"success"`
       Data    interface{} `json:"data"`
       Meta    PaginationMeta `json:"meta"`
   }

   type PaginationMeta struct {
       Cursor  string `json:"cursor,omitempty"`
       HasMore bool   `json:"hasMore"`
       Total   int    `json:"total,omitempty"`
   }

   func JSON(w http.ResponseWriter, status int, data interface{})
   func Error(w http.ResponseWriter, err *errors.AppError)
   func Paginated(w http.ResponseWriter, data interface{}, meta PaginationMeta)
   ```
3. Buat `internal/handler/helpers.go`:
   - `DecodeJSON(r, v) error` — decode request body
   - `GetUserID(ctx) uuid.UUID` — extract user ID dari context
   - `GetPathParam(r, key) string` — extract dari chi URL params
   - `ParsePagination(r) (cursor string, limit int)` — parse query params

### Acceptance Criteria:
- [x] Semua response consistent format (success/error)
- [x] Error codes standardized
- [x] Request body validation
- [x] Pagination support (cursor-based)
- [x] User ID extractable dari context

### Testing:
- [x] Unit test: JSON response formatting
- [x] Unit test: error response formatting
- [x] Unit test: request body decoding
- [x] Unit test: pagination parsing

---

## Task 3.3: WebSocket Hub

**Input:** Task 3.1 selesai
**Output:** WebSocket hub untuk real-time messaging

### Steps:
1. Buat `internal/ws/hub.go`:
   ```go
   type Hub struct {
       clients    map[uuid.UUID]*Client  // userID → client
       rooms      map[string]map[uuid.UUID]*Client  // roomID → clients
       register   chan *Client
       unregister chan *Client
       broadcast  chan *BroadcastMessage
       mu         sync.RWMutex
   }

   type Client struct {
       UserID uuid.UUID
       Conn   *websocket.Conn
       Send   chan []byte
       Hub    *Hub
   }

   type BroadcastMessage struct {
       Room    string
       Data    []byte
       Exclude uuid.UUID // exclude sender
   }

   func NewHub() *Hub
   func (h *Hub) Run()
   func (h *Hub) RegisterClient(client *Client)
   func (h *Hub) UnregisterClient(client *Client)
   func (h *Hub) JoinRoom(client *Client, roomID string)
   func (h *Hub) LeaveRoom(client *Client, roomID string)
   func (h *Hub) SendToUser(userID uuid.UUID, data []byte)
   func (h *Hub) SendToRoom(roomID string, data []byte, excludeUserID uuid.UUID)
   func (h *Hub) IsOnline(userID uuid.UUID) bool
   ```
2. Buat `internal/ws/client.go`:
   ```go
   func (c *Client) ReadPump()   // goroutine: read messages from WS
   func (c *Client) WritePump()  // goroutine: write messages to WS
   ```
   - ReadPump: parse incoming messages, route to handlers
   - WritePump: send queued messages, handle ping/pong
3. Buat `internal/ws/message.go`:
   ```go
   type WSMessage struct {
       Type    string          `json:"type"`
       Payload json.RawMessage `json:"payload"`
   }

   // Message types
   const (
       WSTypeMessage       = "message"
       WSTypeTyping        = "typing"
       WSTypeOnlineStatus  = "online_status"
       WSTypeReadReceipt   = "read_receipt"
       WSTypeDocUpdate     = "doc_update"
       WSTypeDocLock       = "doc_lock"
       WSTypeNotification  = "notification"
   )
   ```
4. Buat `internal/handler/ws_handler.go`:
   - Upgrade HTTP ke WebSocket
   - Authenticate via query param token: `ws://host/ws?token=xxx`
   - Create Client, register to Hub
   - Auto-join rooms berdasarkan user's chats/topics
5. Room naming conventions:
   - Chat room: `chat:{chatID}`
   - Topic room: `topic:{topicID}`
   - Document room: `doc:{docID}`
   - User room: `user:{userID}` (for direct notifications)

### Acceptance Criteria:
- [x] WebSocket upgrade berfungsi
- [x] Authentication via token
- [x] Client register/unregister
- [x] Room join/leave
- [x] Send to specific user
- [x] Broadcast to room (exclude sender)
- [x] Online status tracking
- [x] Ping/pong keepalive
- [x] Graceful disconnect handling

### Testing:
- [x] Unit test: hub register/unregister
- [x] Unit test: room join/leave
- [x] Unit test: broadcast to room
- [x] Unit test: send to user
- [x] Unit test: online status
- [x] Integration test: WebSocket connection lifecycle

---

## Task 3.4: Dependency Injection & Service Layer Template

**Input:** Task 3.1, 3.2, 3.3 selesai
**Output:** DI container dan service layer foundation

### Steps:
1. Buat `internal/handler/dependencies.go`:
   ```go
   type Dependencies struct {
       Config         *config.Config
       DB             *pgxpool.Pool
       Redis          *redis.Client
       Hub            *ws.Hub

       // Repositories
       UserRepo       repository.UserRepository
       ChatRepo       repository.ChatRepository
       MessageRepo    repository.MessageRepository
       TopicRepo      repository.TopicRepository
       DocumentRepo   repository.DocumentRepository
       BlockRepo      repository.BlockRepository
       EntityRepo     repository.EntityRepository

       // Services
       AuthService    service.AuthService
       UserService    service.UserService
       ChatService    service.ChatService
       // ... more services

       // Handlers
       AuthHandler    *AuthHandler
       UserHandler    *UserHandler
       ChatHandler    *ChatHandler
       // ... more handlers
       WSHandler      *WSHandler
   }

   func NewDependencies(cfg *config.Config, db *pgxpool.Pool, redis *redis.Client) *Dependencies {
       // Wire everything together
   }
   ```
2. Buat service interface templates di `internal/service/`:
   - `auth_service.go`: interface + struct placeholder
   - `user_service.go`: interface + struct placeholder
   - `chat_service.go`: interface + struct placeholder
3. Update `cmd/server/main.go` untuk menggunakan DI:
   ```go
   func main() {
       cfg := config.Load()
       db := database.NewPool(cfg.DatabaseURL)
       redis := database.NewRedisClient(cfg.RedisURL)
       database.RunMigrations(cfg.DatabaseURL, "migrations")

       deps := handler.NewDependencies(cfg, db, redis)

       hub := ws.NewHub()
       go hub.Run()

       router := handler.NewRouter(cfg, deps)
       server := &http.Server{Addr: ":" + cfg.Port, Handler: router}
       server.ListenAndServe()
   }
   ```

### Acceptance Criteria:
- [x] All dependencies wired via NewDependencies
- [x] Repository interfaces defined
- [x] Service interfaces defined
- [x] Handler structs accept services via constructor
- [x] Clean separation: handler → service → repository
- [x] Main.go clean dan minimal

### Testing:
- [x] Integration test: server startup
- [x] Integration test: health check via HTTP
- [x] Integration test: WebSocket connection

---

## Phase 03 Review

### Testing Checklist:
- [x] `make dev` — server berjalan dengan semua middleware
- [x] Health check — `GET /health` return 200
- [x] Auth middleware — protected routes return 401 tanpa token
- [x] CORS — proper headers di response
- [x] Rate limit — return 429 setelah limit
- [x] WebSocket — connection establish, ping/pong
- [x] `go test ./...` — semua test pass

### Review Checklist:
- [x] Router structure sesuai REST conventions
- [x] Error codes sesuai `docs/error-handling.md`
- [x] Middleware order correct (logger first, auth before handlers)
- [x] WebSocket properly authenticated
- [x] No goroutine leaks
- [x] Naming sesuai `docs/naming-conventions.md`
- [x] Commit: `feat(api): implement API router and WebSocket hub`
