# Go Style Guide

> **Go Version:** 1.23+
> **Formatter:** `gofmt` / `goimports`
> **Linter:** `golangci-lint`

---

## General Rules

### Formatting
- Gunakan `gofmt` — tidak ada diskusi tentang formatting
- Gunakan `goimports` untuk auto-manage imports
- Max line width: guidance 100 characters (Go tidak enforce)
- Indentation: tabs (Go standard)

### Linting
- Run `golangci-lint run` sebelum commit
- Zero warnings policy
- Enabled linters: `govet`, `errcheck`, `staticcheck`, `gosimple`, `ineffassign`, `unused`, `misspell`

---

## Naming

| Item | Convention | Example |
|------|-----------|---------|
| Package | lowercase, short, no underscore | `auth`, `chat`, `ws` |
| Exported type | PascalCase | `ChatMessage`, `UserProfile` |
| Unexported type | camelCase | `wsClient`, `otpData` |
| Interface | PascalCase, -er suffix | `MessageSender`, `ChatRepository` |
| Exported function | PascalCase | `NewChatService`, `SendMessage` |
| Unexported function | camelCase | `validateOTP`, `buildQuery` |
| Constant | PascalCase (exported) | `MaxGroupMembers`, `OTPLength` |
| Constant | camelCase (unexported) | `defaultTimeout`, `otpExpiry` |
| Error var | `Err` prefix | `ErrNotFound`, `ErrUnauthorized` |
| Method receiver | 1-2 chars | `func (s *Service)`, `func (r *Repo)` |

### Naming Patterns

```go
// Constructor: New{Type}
func NewChatService(repo ChatRepository) *ChatService { ... }

// Interface: verb-er or descriptive
type MessageSender interface {
    Send(ctx context.Context, msg *Message) error
}

type ChatRepository interface {
    FindByID(ctx context.Context, id string) (*Chat, error)
    Create(ctx context.Context, chat *Chat) error
}

// Boolean: Is/Has/Can
func (u *User) IsOnline() bool { ... }
func (d *Document) HasSignature(userID string) bool { ... }
func (m *Member) CanEdit() bool { ... }

// Conversion: To{Type}
func (m *Message) ToResponse() MessageResponse { ... }
func (u *User) ToSummary() UserSummary { ... }
```

---

## Structs

### Data Structs (Models)

```go
// internal/chat/model.go

type Chat struct {
    ID        string    `json:"id" db:"id"`
    Type      ChatType  `json:"type" db:"type"`
    Name      *string   `json:"name,omitempty" db:"name"`
    Icon      *string   `json:"icon,omitempty" db:"icon"`
    MemberIDs []string  `json:"memberIds" db:"-"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type Message struct {
    ID        string    `json:"id" db:"id"`
    ChatID    string    `json:"chatId" db:"chat_id"`
    SenderID  string    `json:"senderId" db:"sender_id"`
    Text      string    `json:"text" db:"text"`
    ReplyTo   *string   `json:"replyTo,omitempty" db:"reply_to"`
    CreatedAt time.Time `json:"createdAt" db:"created_at"`
}
```

### Request/Response DTOs

```go
// Requests: {Action}{Entity}Request
type SendMessageRequest struct {
    Text    string  `json:"text" binding:"required,max=4096"`
    ReplyTo *string `json:"replyTo,omitempty"`
}

type CreateGroupRequest struct {
    Name      string   `json:"name" binding:"required,min=1,max=100"`
    Icon      string   `json:"icon" binding:"required"`
    MemberIDs []string `json:"memberIds" binding:"required,min=2"`
}

// Responses: {Entity}Response (when different from model)
type ChatSummary struct {
    ID          string   `json:"id"`
    Type        ChatType `json:"type"`
    Name        string   `json:"name"`
    LastMessage *Message `json:"lastMessage,omitempty"`
    UnreadCount int      `json:"unreadCount"`
}

// Validation method on request
func (r *SendMessageRequest) Validate() error {
    if strings.TrimSpace(r.Text) == "" {
        return errors.New("message text cannot be empty")
    }
    return nil
}
```

---

## Enums (Typed Constants)

```go
// internal/chat/model.go

type ChatType string

const (
    ChatTypePersonal ChatType = "personal"
    ChatTypeGroup    ChatType = "group"
)

func (ct ChatType) IsValid() bool {
    switch ct {
    case ChatTypePersonal, ChatTypeGroup:
        return true
    }
    return false
}

// Block types for documents
type BlockType string

const (
    BlockParagraph  BlockType = "paragraph"
    BlockHeading1   BlockType = "heading1"
    BlockHeading2   BlockType = "heading2"
    BlockHeading3   BlockType = "heading3"
    BlockBulletList BlockType = "bullet-list"
    BlockNumberList BlockType = "numbered-list"
    BlockChecklist  BlockType = "checklist"
    BlockTable      BlockType = "table"
    BlockCallout    BlockType = "callout"
    BlockCode       BlockType = "code"
    BlockToggle     BlockType = "toggle"
    BlockDivider    BlockType = "divider"
    BlockQuote      BlockType = "quote"
)
```

---

## Error Handling

```go
// Always check errors
chat, err := s.repo.FindByID(ctx, chatID)
if err != nil {
    return nil, fmt.Errorf("find chat: %w", err)
}

// Use error wrapping with %w for chain
func (s *Service) SendMessage(ctx context.Context, msg *Message) error {
    if err := s.repo.Save(ctx, msg); err != nil {
        return fmt.Errorf("send message: %w", err)
    }
    return nil
}

// Use errors.Is / errors.As for checking
if errors.Is(err, apperror.ErrNotFound) {
    // handle not found
}

var appErr *apperror.AppError
if errors.As(err, &appErr) {
    // handle typed error
}

// NEVER use panic in handlers/services
// Only acceptable in main() init or truly unrecoverable situations
```

---

## Context

```go
// Always pass context as first parameter
func (s *Service) GetChat(ctx context.Context, id string) (*Chat, error) { ... }

// Use context for cancellation and timeouts
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()

result, err := s.repo.FindByID(ctx, id)

// Don't store context in structs
// WRONG:
type Service struct {
    ctx context.Context  // Never do this
}

// RIGHT: Pass context per-method
func (s *Service) DoWork(ctx context.Context) error { ... }
```

---

## Concurrency

```go
// Use goroutines for async work
go s.notifier.NotifyChatMessage(context.Background(), chat, msg)

// Use channels for communication
results := make(chan Result, workerCount)
for i := 0; i < workerCount; i++ {
    go func() {
        results <- doWork()
    }()
}

// Use sync.Mutex for shared state
type Hub struct {
    clients map[string]*Client
    mu      sync.RWMutex
}

func (h *Hub) GetClient(id string) *Client {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.clients[id]
}

func (h *Hub) AddClient(id string, c *Client) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.clients[id] = c
}

// Use errgroup for concurrent operations with error handling
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return fetchUsers(ctx) })
g.Go(func() error { return fetchChats(ctx) })
if err := g.Wait(); err != nil {
    return err
}
```

---

## HTTP Handlers (Gin)

```go
// Handler methods follow RESTful conventions
func (h *ChatHandler) GetChat(c *gin.Context) {
    chatID := c.Param("id")
    userID := middleware.GetUserID(c)

    chat, err := h.service.GetChat(c.Request.Context(), userID, chatID)
    if err != nil {
        handleError(c, err)
        return
    }

    response.OK(c, chat)
}

// Route registration in one place
func (h *ChatHandler) RegisterRoutes(r *gin.RouterGroup, auth gin.HandlerFunc) {
    chats := r.Group("/chats", auth)
    {
        chats.GET("", h.ListChats)
        chats.POST("", h.CreateChat)
        chats.GET("/:id", h.GetChat)
        chats.POST("/:id/messages", h.SendMessage)
        chats.GET("/:id/messages", h.GetMessages)
    }
}
```

---

## Module Organization

```go
// Each domain package has clear public interface
// internal/chat/

// model.go     — Data types
// service.go   — Business logic (exported)
// repository.go — Data access interface (exported interface)
// handler.go   — HTTP handlers (exported)
// hub.go       — WebSocket for this domain (if needed)

// Keep implementation details unexported
// Only export what other packages need
```

---

## Testing

```go
// Co-located tests in same package
// internal/chat/service_test.go

func TestChatService_SendMessage(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Arrange
        repo := new(MockChatRepository)
        svc := NewChatService(repo, nil, nil)

        // Act
        msg, err := svc.SendMessage(ctx, "user-1", "chat-1", req)

        // Assert
        assert.NoError(t, err)
        assert.Equal(t, "Hello", msg.Text)
    })
}

// Use table-driven tests for multiple scenarios
// Use testify for assertions
// Use interfaces + mocks for dependencies
```

---

## Documentation

```go
// Package-level comment
// Package chat implements the chat functionality including
// personal chats, group chats, and message delivery.
package chat

// Exported type/function doc comments
// ChatService handles all chat-related business logic including
// message sending, chat creation, and member management.
type ChatService struct { ... }

// SendMessage sends a message to a chat, broadcasting it to all
// online members via WebSocket and sending push notifications
// to offline members.
func (s *ChatService) SendMessage(ctx context.Context, ...) (*Message, error) { ... }
```

---

## Performance Guidelines

1. **String handling:** Use `strings.Builder` for concatenation in loops
2. **Slices:** Pre-allocate with `make([]T, 0, capacity)` when size is known
3. **Maps:** Pre-allocate with `make(map[K]V, capacity)` when size is known
4. **Database:** Use connection pooling, prepare statements for repeated queries
5. **JSON:** Use `json.NewEncoder`/`json.NewDecoder` for streaming
6. **Concurrency:** Use `sync.Pool` for frequently allocated objects
7. **Memory:** Stream large files instead of loading fully into memory
8. **Goroutines:** Don't leak — always ensure goroutines can be cancelled via context

---

## Forbidden Practices

| Practice | Why | Alternative |
|----------|-----|------------|
| `panic()` in handlers | App crash | Return `error` |
| Naked `return` in complex functions | Hard to read | Named returns or explicit |
| `init()` for complex logic | Order-dependent | Explicit initialization |
| Global mutable state | Race conditions | Dependency injection |
| `interface{}` / `any` everywhere | No type safety | Concrete types or generics |
| `log.Fatal()` in handlers | App exit mid-request | Log + return error |
| Ignoring errors (`_ = fn()`) | Hidden bugs | Handle or log with comment |
| Context in struct fields | Misuse | Pass per-method |
