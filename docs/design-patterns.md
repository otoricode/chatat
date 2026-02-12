# Design Patterns

> Pola desain yang digunakan di backend (Go) dan frontend (React Native/TypeScript).
> Setiap pattern disertai contoh implementasi spesifik untuk Chatat.

---

## Backend (Go) Patterns

### 1. Repository Pattern

**Tujuan:** Abstraksi akses data / database.

```go
// internal/chat/repository.go

type ChatRepository interface {
    FindByID(ctx context.Context, id string) (*Chat, error)
    Create(ctx context.Context, chat *Chat) error
    Update(ctx context.Context, chat *Chat) error
    Delete(ctx context.Context, id string) error
    ListByUser(ctx context.Context, userID string) ([]*ChatSummary, error)
    SaveMessage(ctx context.Context, chatID string, msg *Message) error
    GetMessages(ctx context.Context, chatID string, cursor string, limit int) ([]*Message, error)
}

type pgChatRepository struct {
    db *pgxpool.Pool
}

func NewChatRepository(db *pgxpool.Pool) ChatRepository {
    return &pgChatRepository{db: db}
}

func (r *pgChatRepository) FindByID(ctx context.Context, id string) (*Chat, error) {
    var chat Chat
    err := r.db.QueryRow(ctx,
        `SELECT id, type, name, icon, created_at FROM chats WHERE id = $1`, id,
    ).Scan(&chat.ID, &chat.Type, &chat.Name, &chat.Icon, &chat.CreatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, apperror.NotFound("chat", id)
        }
        return nil, fmt.Errorf("find chat by id: %w", err)
    }
    return &chat, nil
}

func (r *pgChatRepository) SaveMessage(ctx context.Context, chatID string, msg *Message) error {
    _, err := r.db.Exec(ctx,
        `INSERT INTO messages (id, chat_id, sender_id, text, reply_to, created_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
        msg.ID, chatID, msg.SenderID, msg.Text, msg.ReplyTo, msg.CreatedAt,
    )
    if err != nil {
        return fmt.Errorf("save message: %w", err)
    }
    return nil
}
```

**Aturan:**
- Satu repository per domain entity
- Repository hanya berisi operasi CRUD + query
- Tidak ada business logic di repository
- Selalu gunakan interface untuk testability
- Selalu terima `context.Context` sebagai parameter pertama

---

### 2. Service Pattern

**Tujuan:** Business logic layer antara handler dan repository.

```go
// internal/chat/service.go

type ChatService struct {
    chatRepo    ChatRepository
    userRepo    user.UserRepository
    wsHub       *ws.Hub
    notifier    *notification.Service
}

func NewChatService(
    chatRepo ChatRepository,
    userRepo user.UserRepository,
    wsHub *ws.Hub,
    notifier *notification.Service,
) *ChatService {
    return &ChatService{
        chatRepo: chatRepo,
        userRepo: userRepo,
        wsHub:    wsHub,
        notifier: notifier,
    }
}

func (s *ChatService) SendMessage(ctx context.Context, senderID, chatID string, req SendMessageRequest) (*Message, error) {
    // 1. Validate sender is member of chat
    chat, err := s.chatRepo.FindByID(ctx, chatID)
    if err != nil {
        return nil, err
    }

    if !chat.IsMember(senderID) {
        return nil, apperror.Forbidden("you are not a member of this chat")
    }

    // 2. Create message
    msg := &Message{
        ID:        uuid.NewString(),
        ChatID:    chatID,
        SenderID:  senderID,
        Text:      req.Text,
        ReplyTo:   req.ReplyTo,
        CreatedAt: time.Now(),
    }

    // 3. Persist
    if err := s.chatRepo.SaveMessage(ctx, chatID, msg); err != nil {
        return nil, err
    }

    // 4. Broadcast via WebSocket
    s.wsHub.BroadcastToChat(chatID, ws.Event{
        Type:    "chat:message_sent",
        Payload: msg,
    })

    // 5. Push notification to offline members
    go s.notifier.NotifyChatMessage(context.Background(), chat, msg, senderID)

    return msg, nil
}
```

**Aturan:**
- Service orchestrate repository + infrastructure
- Service boleh memanggil service lain
- Gunakan dependency injection via constructor
- Business logic validasi ada di service, bukan handler

---

### 3. Handler Pattern (HTTP Layer)

**Tujuan:** Thin layer yang menerima request, delegate ke service, return response.

```go
// internal/chat/handler.go

type ChatHandler struct {
    service *ChatService
}

func NewChatHandler(service *ChatService) *ChatHandler {
    return &ChatHandler{service: service}
}

func (h *ChatHandler) RegisterRoutes(r *gin.RouterGroup, authMw gin.HandlerFunc) {
    chats := r.Group("/chats", authMw)
    {
        chats.GET("", h.ListChats)
        chats.POST("", h.CreateChat)
        chats.GET("/:id", h.GetChat)
        chats.POST("/:id/messages", h.SendMessage)
        chats.GET("/:id/messages", h.GetMessages)
    }
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
    userID := middleware.GetUserID(c)
    chatID := c.Param("id")

    var req SendMessageRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, apperror.BadRequest("invalid request body"))
        return
    }

    if err := req.Validate(); err != nil {
        response.Error(c, apperror.BadRequest(err.Error()))
        return
    }

    msg, err := h.service.SendMessage(c.Request.Context(), userID, chatID, req)
    if err != nil {
        var appErr *apperror.AppError
        if errors.As(err, &appErr) {
            response.Error(c, appErr)
        } else {
            response.Error(c, apperror.Internal(err))
        }
        return
    }

    response.Created(c, msg)
}
```

**Aturan:**
- Thin layer — parse request, delegate, return response
- Semua business logic di service
- Error conversion ke HTTP ada di handler
- Validasi input di handler sebelum call service

---

### 4. WebSocket Hub Pattern

**Tujuan:** Manage WebSocket connections dan message broadcasting.

```go
// internal/ws/hub.go

type Hub struct {
    clients    map[string]*Client  // userID -> Client
    register   chan *Client
    unregister chan *Client
    mu         sync.RWMutex
}

type Client struct {
    UserID string
    Conn   *websocket.Conn
    Send   chan []byte
}

type Event struct {
    Type    string      `json:"type"`
    Payload interface{} `json:"payload"`
}

func NewHub() *Hub {
    return &Hub{
        clients:    make(map[string]*Client),
        register:   make(chan *Client),
        unregister: make(chan *Client),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client.UserID] = client
            h.mu.Unlock()

        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client.UserID]; ok {
                delete(h.clients, client.UserID)
                close(client.Send)
            }
            h.mu.Unlock()
        }
    }
}

func (h *Hub) SendToUser(userID string, event Event) {
    h.mu.RLock()
    client, ok := h.clients[userID]
    h.mu.RUnlock()

    if ok {
        data, _ := json.Marshal(event)
        select {
        case client.Send <- data:
        default:
            // Client buffer full, skip
        }
    }
}

func (h *Hub) BroadcastToChat(chatID string, event Event, memberIDs []string) {
    for _, memberID := range memberIDs {
        h.SendToUser(memberID, event)
    }
}
```

---

### 5. Middleware Pattern

**Tujuan:** Cross-cutting concerns (auth, logging, rate limiting).

```go
// internal/middleware/auth.go

func AuthRequired(tokenService *auth.TokenService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractBearerToken(c.GetHeader("Authorization"))
        if token == "" {
            response.Error(c, apperror.Unauthorized("missing token"))
            c.Abort()
            return
        }

        claims, err := tokenService.Validate(token)
        if err != nil {
            response.Error(c, apperror.Unauthorized("invalid or expired token"))
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Next()
    }
}

func GetUserID(c *gin.Context) string {
    return c.GetString("user_id")
}
```

---

### 6. Options Pattern

**Tujuan:** Konfigurasi yang fleksibel dan opsional.

```go
// internal/config/config.go

type ServerConfig struct {
    Port         int
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    MaxBodySize  int64
}

type Option func(*ServerConfig)

func WithPort(port int) Option {
    return func(c *ServerConfig) { c.Port = port }
}

func WithReadTimeout(d time.Duration) Option {
    return func(c *ServerConfig) { c.ReadTimeout = d }
}

func NewServerConfig(opts ...Option) *ServerConfig {
    cfg := &ServerConfig{
        Port:         8080,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        MaxBodySize:  10 << 20, // 10MB
    }
    for _, opt := range opts {
        opt(cfg)
    }
    return cfg
}

// Usage:
cfg := NewServerConfig(
    WithPort(3000),
    WithReadTimeout(30 * time.Second),
)
```

---

## Frontend (React Native) Patterns

### 1. Screen/Component Separation

**Tujuan:** Screen = data fetching + layout. Component = pure UI.

```typescript
// screens/chat/ChatScreen.tsx (Container)
export function ChatScreen({ route }: ChatScreenProps) {
  const { chatId } = route.params;
  const { messages, isLoading, error, sendMessage } = useChat(chatId);
  const { user } = useAuth();

  if (isLoading) return <LoadingState />;
  if (error) return <ErrorState error={error} />;

  return (
    <ChatView
      messages={messages}
      currentUserId={user.id}
      onSend={sendMessage}
    />
  );
}

// components/chat/ChatView.tsx (Presenter - pure UI)
interface ChatViewProps {
  messages: Message[];
  currentUserId: string;
  onSend: (text: string) => void;
}

export function ChatView({ messages, currentUserId, onSend }: ChatViewProps) {
  return (
    <View style={styles.container}>
      <FlatList
        data={messages}
        renderItem={({ item }) => (
          <MessageBubble
            message={item}
            isMine={item.senderId === currentUserId}
          />
        )}
        inverted
      />
      <MessageInput onSend={onSend} />
    </View>
  );
}
```

---

### 2. Custom Hook Pattern

**Tujuan:** Encapsulate data fetching + state management per domain.

```typescript
// hooks/useDocument.ts

export function useDocument(documentId?: string) {
  const [document, setDocument] = useState<Document | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);

  const fetchDocument = useCallback(async () => {
    if (!documentId) return;
    setIsLoading(true);
    setError(null);
    try {
      const doc = await documentApi.getById(documentId);
      setDocument(doc);
    } catch (err) {
      setError(parseError(err));
    } finally {
      setIsLoading(false);
    }
  }, [documentId]);

  useEffect(() => { fetchDocument(); }, [fetchDocument]);

  const lockDocument = useCallback(async () => {
    if (!documentId) return;
    try {
      await documentApi.lock(documentId);
      await fetchDocument(); // Refresh
    } catch (err) {
      setError(parseError(err));
    }
  }, [documentId, fetchDocument]);

  return { document, isLoading, error, refresh: fetchDocument, lockDocument };
}
```

---

### 3. Store Pattern (Zustand)

**Tujuan:** Global state management ringan dan predictable.

```typescript
// stores/chatStore.ts

import { create } from 'zustand';

interface ChatState {
  chats: ChatSummary[];
  activeChatId: string | null;
  unreadCounts: Record<string, number>;

  setChats: (chats: ChatSummary[]) => void;
  setActiveChat: (id: string | null) => void;
  incrementUnread: (chatId: string) => void;
  clearUnread: (chatId: string) => void;
  updateLastMessage: (chatId: string, message: Message) => void;
}

export const useChatStore = create<ChatState>((set) => ({
  chats: [],
  activeChatId: null,
  unreadCounts: {},

  setChats: (chats) => set({ chats }),
  setActiveChat: (id) => set({ activeChatId: id }),

  incrementUnread: (chatId) =>
    set((state) => ({
      unreadCounts: {
        ...state.unreadCounts,
        [chatId]: (state.unreadCounts[chatId] || 0) + 1,
      },
    })),

  clearUnread: (chatId) =>
    set((state) => ({
      unreadCounts: { ...state.unreadCounts, [chatId]: 0 },
    })),

  updateLastMessage: (chatId, message) =>
    set((state) => ({
      chats: state.chats.map((chat) =>
        chat.id === chatId
          ? { ...chat, lastMessage: message, updatedAt: message.createdAt }
          : chat
      ),
    })),
}));
```

---

### 4. WebSocket Integration Pattern

**Tujuan:** Real-time events dari server ke React Native.

```typescript
// hooks/useWebSocket.ts

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null);
  const { token } = useAuth();
  const chatStore = useChatStore();

  useEffect(() => {
    if (!token) return;

    const ws = new WebSocket(`wss://api.chatat.app/ws?token=${token}`);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      const { type, payload } = JSON.parse(event.data);

      switch (type) {
        case 'chat:message_sent':
          chatStore.updateLastMessage(payload.chatId, payload);
          if (chatStore.activeChatId !== payload.chatId) {
            chatStore.incrementUnread(payload.chatId);
          }
          break;

        case 'document:locked':
          documentStore.getState().markLocked(payload.documentId);
          break;

        case 'presence:online':
          contactStore.getState().setOnline(payload.userId);
          break;
      }
    };

    ws.onclose = () => {
      // Auto-reconnect after delay
      setTimeout(() => reconnect(), 3000);
    };

    return () => ws.close();
  }, [token]);
}
```

---

## Pattern Decision Matrix

| Situation | Pattern | Layer |
|-----------|---------|-------|
| Database operations | Repository (interface) | Backend |
| Business logic orchestration | Service | Backend |
| HTTP request handling | Handler (thin) | Backend |
| Real-time messaging | WebSocket Hub | Backend |
| Cross-cutting concerns | Middleware | Backend |
| Flexible configuration | Options pattern | Backend |
| Data fetch + rendering | Screen/Component separation | Frontend |
| Domain data management | Custom Hook | Frontend |
| Global app state | Zustand Store | Frontend |
| Real-time updates | WebSocket hook | Frontend |
| Navigation | React Navigation | Frontend |
| Form + validation | React Hook Form + Zod | Frontend |
