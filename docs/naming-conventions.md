# Naming Conventions

> Konvensi penamaan lengkap untuk seluruh codebase Chatat.
> Referensi tunggal — semua aturan naming terkumpul di sini.

---

## File & Folder Naming

### Backend (Go / server/)

| Item | Convention | Example |
|------|-----------|---------|
| Package folder | lowercase, no underscore | `auth/`, `chat/`, `document/` |
| Source file | snake_case.go | `reverse_otp.go`, `hub.go` |
| Test file | `*_test.go` | `service_test.go` |
| Config file | lowercase | `go.mod`, `Dockerfile` |
| Migration | `NNN_description.sql` | `001_create_users.sql` |

### Frontend (React Native / mobile/)

| Item | Convention | Example |
|------|-----------|---------|
| Screen file | PascalCase + Screen | `ChatScreen.tsx` |
| Component file | PascalCase.tsx | `MessageBubble.tsx` |
| Hook file | camelCase.ts | `useChat.ts` |
| Store file | camelCase.ts | `chatStore.ts` |
| Service/API file | camelCase.ts | `chatApi.ts` |
| Type file | camelCase.ts | `chat.ts` |
| Util file | camelCase.ts | `formatting.ts` |
| Constant file | camelCase.ts | `constants.ts` |
| Test file | `*.test.ts(x)` | `ChatScreen.test.tsx` |
| Locale file | ISO 639-1 | `id.json`, `en.json`, `ar.json` |

---

## Go Naming

### Identifiers

| Item | Convention | Example |
|------|-----------|---------|
| Package | lowercase, short | `auth`, `chat`, `ws` |
| Exported struct | PascalCase | `ChatMessage` |
| Unexported struct | camelCase | `wsClient` |
| Interface | PascalCase, -er suffix | `MessageSender`, `Repository` |
| Exported function | PascalCase | `SendMessage` |
| Unexported function | camelCase | `validateOTP` |
| Exported variable | PascalCase | `MaxRetries` |
| Unexported variable | camelCase | `defaultTimeout` |
| Constant (exported) | PascalCase | `MaxGroupMembers` |
| Constant (unexported) | camelCase | `otpLength` |
| Enum-like const | PascalCase | `ChatTypePersonal`, `ChatTypeGroup` |
| Error variable | `Err` prefix | `ErrNotFound`, `ErrUnauthorized` |
| Context key | unexported type | `type ctxKey string` |
| Method receiver | 1-2 char lowercase | `func (s *Service)`, `func (r *Repository)` |

### Common Patterns

```go
// Constructor: New{Type}
func NewChatService(repo ChatRepository) *ChatService { ... }

// Interface naming: verb-er
type MessageSender interface {
    SendMessage(ctx context.Context, msg *Message) error
}

type ChatRepository interface {
    FindByID(ctx context.Context, id string) (*Chat, error)
    Create(ctx context.Context, chat *Chat) error
    Update(ctx context.Context, chat *Chat) error
    Delete(ctx context.Context, id string) error
    ListByUser(ctx context.Context, userID string) ([]*Chat, error)
}

// Error variables
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrAlreadyExists = errors.New("already exists")
)

// Boolean functions: Is/Has/Can prefix
func (u *User) IsOnline() bool { ... }
func (d *Document) HasSignedAll() bool { ... }
func (m *Member) CanEdit() bool { ... }

// Options pattern
type Option func(*Config)
func WithTimeout(d time.Duration) Option { ... }
func WithMaxRetries(n int) Option { ... }
```

### Module & Type Naming

```go
// Repository: {Entity}Repository
ChatRepository, UserRepository, DocumentRepository

// Service: {Entity}Service
ChatService, AuthService, DocumentService

// Handler: {Entity}Handler
ChatHandler, AuthHandler, DocumentHandler

// Model: domain name directly
User, Chat, Message, Document, Block, Topic, Entity

// DTO: {Action}{Entity}Request / {Entity}Response
CreateChatRequest, SendMessageRequest
ChatResponse, MessageResponse

// Event payload: {Entity}{Event}
MessageSent, DocumentLocked, TopicCreated
```

---

## TypeScript Naming

### Identifiers

| Item | Convention | Example |
|------|-----------|---------|
| Screen component | PascalCase + Screen | `ChatScreen` |
| Component | PascalCase | `MessageBubble` |
| Hook | camelCase, `use` prefix | `useChat` |
| Function | camelCase | `formatTimestamp` |
| Variable | camelCase | `messageCount` |
| Constant | SCREAMING_SNAKE | `MAX_MESSAGE_LENGTH` |
| Type | PascalCase | `ChatMessage` |
| Interface | PascalCase | `ChatRepository` |
| Enum | PascalCase | `ChatType` |
| Enum member | PascalCase | `ChatType.Personal` |
| Store | camelCase, `Store` suffix | `chatStore` |
| Context | PascalCase, `Context` suffix | `AuthContext` |
| Provider | PascalCase, `Provider` suffix | `AuthProvider` |

### Common Prefixes & Suffixes

| Pattern | Usage | Example |
|---------|-------|---------|
| `use*` | React hook | `useChat`, `useDocument` |
| `handle*` | Event handler | `handleSend`, `handleLongPress` |
| `on*` | Callback prop | `onSend`, `onSelect` |
| `is*` | Boolean state | `isLoading`, `isLocked` |
| `has*` | Boolean existence | `hasUnread`, `hasSigned` |
| `can*` | Boolean permission | `canEdit`, `canLock` |
| `*Screen` | Screen component | `ChatScreen`, `LoginScreen` |
| `*Store` | Zustand store | `chatStore` |
| `*Api` | API service | `chatApi` |
| `*Db` | Local DB module | `chatDb` |
| `*Props` | Component props | `MessageBubbleProps` |

### Screen Naming

```
// Auth screens: {Action}Screen
LoginScreen, OTPScreen, ProfileSetupScreen

// Chat screens: Chat{Descriptor}Screen
ChatListScreen, ChatScreen, ChatInfoScreen

// Document screens: Document{Descriptor}Screen
DocumentListScreen, DocumentScreen, DocumentInfoScreen

// Topic screens: Topic{Descriptor}Screen
TopicListScreen, TopicScreen, NewTopicScreen

// Settings: {Feature}Screen
SettingsScreen, ProfileScreen, LanguageScreen
```

---

## Database Naming (PostgreSQL)

| Item | Convention | Example |
|------|-----------|---------|
| Table | snake_case, plural | `users`, `chat_messages` |
| Column | snake_case | `created_at`, `sender_id` |
| Primary key | `id` | `id UUID PRIMARY KEY` |
| Foreign key | `{entity}_id` | `chat_id`, `user_id` |
| Boolean column | `is_*` or `has_*` | `is_locked`, `has_signed` |
| Timestamp | `*_at` | `created_at`, `locked_at` |
| JSON column | descriptive name | `blocks`, `metadata` |
| Index | `idx_{table}_{columns}` | `idx_messages_chat_id` |
| Unique constraint | `uq_{table}_{columns}` | `uq_users_phone` |
| Enum type | snake_case | `chat_type`, `block_type` |

---

## API Endpoint Naming

### Convention: REST, plural nouns, kebab-case

```
POST   /api/v1/auth/otp/request
POST   /api/v1/auth/otp/verify
POST   /api/v1/auth/reverse-otp/verify

GET    /api/v1/chats
POST   /api/v1/chats
GET    /api/v1/chats/:id
DELETE /api/v1/chats/:id

POST   /api/v1/chats/:id/messages
GET    /api/v1/chats/:id/messages

POST   /api/v1/groups
GET    /api/v1/groups/:id
PUT    /api/v1/groups/:id
POST   /api/v1/groups/:id/members

POST   /api/v1/topics
GET    /api/v1/topics/:id
GET    /api/v1/topics?parent_type=group&parent_id=xxx

POST   /api/v1/documents
GET    /api/v1/documents/:id
PUT    /api/v1/documents/:id
POST   /api/v1/documents/:id/lock
POST   /api/v1/documents/:id/sign

POST   /api/v1/contacts/sync
GET    /api/v1/contacts

POST   /api/v1/media/upload
GET    /api/v1/media/:id
```

---

## WebSocket Event Naming

### Convention: `{domain}:{action}` (snake_case after colon)

```
// Chat events
chat:message_sent
chat:message_delivered
chat:message_read
chat:typing_start
chat:typing_stop

// Group events
group:member_added
group:member_removed
group:info_updated

// Topic events
topic:created
topic:message_sent

// Document events
document:block_updated
document:locked
document:signed
document:collaborator_joined

// Presence events
presence:online
presence:offline
presence:last_seen
```

---

## Git Naming

| Item | Convention | Example |
|------|-----------|---------|
| Branch: feature | `feature/{short-desc}` | `feature/document-locking` |
| Branch: bugfix | `fix/{short-desc}` | `fix/otp-timeout` |
| Branch: refactor | `refactor/{short-desc}` | `refactor/ws-hub` |
| Branch: release | `release/{version}` | `release/1.0.0` |
| Tag | `v{semver}` | `v1.0.0` |
| Commit | Conventional Commits | `feat(chat): add typing indicator` |
