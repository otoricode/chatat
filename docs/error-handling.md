# Error Handling Strategy

> Strategi error handling di seluruh stack: Go backend dan React Native frontend.
> Prinsip: errors as values, bukan exceptions.

---

## Prinsip Utama

1. **Errors as Values** — Gunakan `error` return di Go, typed errors di TypeScript
2. **No Panic** — Jangan pernah `panic()` di production code kecuali init
3. **Error Chain** — Setiap layer menambah konteks dengan `fmt.Errorf("context: %w", err)`
4. **User-Friendly** — Error yang sampai ke client harus readable, bukan stack trace
5. **Recoverable vs Fatal** — Bedakan error yang bisa di-retry vs yang harus stop

---

## Backend (Go)

### Error Type Hierarchy

```
AppError (root)
├── AuthError           — OTP invalid, token expired, unauthorized
├── ChatError           — Message send failed, chat not found
├── DocumentError       — Lock conflict, collaboration error
├── TopicError          — Topic not found, permission denied
├── StorageError        — File upload/download failed
├── DatabaseError       — Query failed, constraint violation
├── ValidationError     — Input validation failures
├── WebSocketError      — Connection dropped, message parse error
└── NotificationError   — Push notification delivery failed
```

### Definisi Error

```go
// pkg/apperror/errors.go

package apperror

import (
    "errors"
    "fmt"
    "net/http"
)

// Sentinel errors
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrConflict      = errors.New("conflict")
    ErrBadRequest    = errors.New("bad request")
    ErrInternal      = errors.New("internal error")
    ErrAlreadyExists = errors.New("already exists")
)

// AppError wraps errors with HTTP status and user message
type AppError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    HTTPStatus int    `json:"-"`
    Err        error  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructors
func NotFound(entity, id string) *AppError {
    return &AppError{
        Code:       "NOT_FOUND",
        Message:    fmt.Sprintf("%s '%s' not found", entity, id),
        HTTPStatus: http.StatusNotFound,
        Err:        ErrNotFound,
    }
}

func Unauthorized(msg string) *AppError {
    return &AppError{
        Code:       "UNAUTHORIZED",
        Message:    msg,
        HTTPStatus: http.StatusUnauthorized,
        Err:        ErrUnauthorized,
    }
}

func Forbidden(msg string) *AppError {
    return &AppError{
        Code:       "FORBIDDEN",
        Message:    msg,
        HTTPStatus: http.StatusForbidden,
        Err:        ErrForbidden,
    }
}

func BadRequest(msg string) *AppError {
    return &AppError{
        Code:       "BAD_REQUEST",
        Message:    msg,
        HTTPStatus: http.StatusBadRequest,
        Err:        ErrBadRequest,
    }
}

func Conflict(msg string) *AppError {
    return &AppError{
        Code:       "CONFLICT",
        Message:    msg,
        HTTPStatus: http.StatusConflict,
        Err:        ErrConflict,
    }
}

func Validation(field, msg string) *AppError {
    return &AppError{
        Code:       "VALIDATION_ERROR",
        Message:    fmt.Sprintf("validation error on '%s': %s", field, msg),
        HTTPStatus: http.StatusUnprocessableEntity,
        Err:        ErrBadRequest,
    }
}

func Internal(err error) *AppError {
    return &AppError{
        Code:       "INTERNAL_ERROR",
        Message:    "an internal error occurred",
        HTTPStatus: http.StatusInternalServerError,
        Err:        err,
    }
}
```

### Error Propagation Rules

```go
// BENAR: Tambah konteks dengan wrapping
func (s *ChatService) SendMessage(ctx context.Context, chatID string, msg *Message) error {
    chat, err := s.repo.FindByID(ctx, chatID)
    if err != nil {
        return fmt.Errorf("send message - find chat: %w", err)
    }

    if err := s.repo.SaveMessage(ctx, chat.ID, msg); err != nil {
        return fmt.Errorf("send message - save: %w", err)
    }

    return nil
}

// BENAR: Convert ke AppError di handler layer
func (h *ChatHandler) SendMessage(c *gin.Context) {
    // ...
    if err := h.service.SendMessage(ctx, chatID, msg); err != nil {
        if errors.Is(err, apperror.ErrNotFound) {
            response.Error(c, apperror.NotFound("chat", chatID))
            return
        }
        response.Error(c, apperror.Internal(err))
        return
    }
    response.OK(c, msg)
}

// SALAH: Jangan ignore error
_ = s.repo.SaveMessage(ctx, chatID, msg) // Error hilang

// SALAH: Jangan pakai panic
panic("unexpected state") // App crash

// BENAR: Log jika intentionally ignoring
if err := s.cache.Invalidate(ctx, key); err != nil {
    slog.Warn("failed to invalidate cache", "key", key, "error", err)
}
```

### HTTP Error Response

```go
// pkg/response/json.go

package response

import (
    "github.com/gin-gonic/gin"
    "github.com/otoritech/chatat/pkg/apperror"
    "log/slog"
)

type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func Error(c *gin.Context, err *apperror.AppError) {
    // Log internal error details
    if err.HTTPStatus >= 500 {
        slog.Error("internal error",
            "code", err.Code,
            "error", err.Err,
            "path", c.Request.URL.Path,
        )
    }

    c.JSON(err.HTTPStatus, ErrorResponse{
        Code:    err.Code,
        Message: err.Message,
    })
}

func OK(c *gin.Context, data any) {
    c.JSON(200, data)
}

func Created(c *gin.Context, data any) {
    c.JSON(201, data)
}
```

### Retry Policy

```go
// pkg/retry/retry.go

package retry

import (
    "context"
    "math"
    "time"
)

type Config struct {
    MaxRetries     int
    InitialDelay   time.Duration
    MaxDelay       time.Duration
    BackoffFactor  float64
}

var Default = Config{
    MaxRetries:    3,
    InitialDelay:  500 * time.Millisecond,
    MaxDelay:      30 * time.Second,
    BackoffFactor: 2.0,
}

func Do(ctx context.Context, cfg Config, fn func() error) error {
    var lastErr error
    delay := cfg.InitialDelay

    for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
        if err := fn(); err != nil {
            lastErr = err
            if attempt < cfg.MaxRetries {
                select {
                case <-time.After(delay):
                case <-ctx.Done():
                    return ctx.Err()
                }
                delay = time.Duration(math.Min(
                    float64(delay)*cfg.BackoffFactor,
                    float64(cfg.MaxDelay),
                ))
            }
        } else {
            return nil
        }
    }

    return lastErr
}
```

---

## Frontend (React Native / TypeScript)

### Error Types

```typescript
// types/api.ts

export interface ApiError {
  code: ErrorCode;
  message: string;
}

export type ErrorCode =
  | 'BAD_REQUEST'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'NOT_FOUND'
  | 'CONFLICT'
  | 'VALIDATION_ERROR'
  | 'INTERNAL_ERROR'
  | 'NETWORK_ERROR';

// Type guard
export function isApiError(error: unknown): error is ApiError {
  return (
    typeof error === 'object' &&
    error !== null &&
    'code' in error &&
    'message' in error
  );
}

// Parse error from API response
export function parseError(error: unknown): ApiError {
  if (isApiError(error)) return error;
  if (error instanceof Error) {
    if (error.message.includes('Network')) {
      return { code: 'NETWORK_ERROR', message: 'No internet connection' };
    }
    return { code: 'INTERNAL_ERROR', message: error.message };
  }
  return { code: 'INTERNAL_ERROR', message: 'An unknown error occurred' };
}
```

### Hook Error Handling

```typescript
// hooks/useChat.ts

export function useChat(chatId?: string) {
  const [error, setError] = useState<ApiError | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const sendMessage = useCallback(async (text: string) => {
    setError(null);
    try {
      await chatApi.sendMessage(chatId, { text });
    } catch (err) {
      const apiError = parseError(err);
      setError(apiError);

      if (apiError.code === 'UNAUTHORIZED') {
        authStore.getState().logout();
      }
    }
  }, [chatId]);

  return { messages, isLoading, error, sendMessage };
}
```

### Error Display

```typescript
// components/shared/ErrorState.tsx

interface ErrorStateProps {
  error: ApiError;
  onRetry?: () => void;
}

export function ErrorState({ error, onRetry }: ErrorStateProps) {
  const title = getErrorTitle(error);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>{title}</Text>
      <Text style={styles.message}>{error.message}</Text>
      {onRetry && (
        <Button onPress={onRetry} title="Try Again" />
      )}
    </View>
  );
}

function getErrorTitle(error: ApiError): string {
  switch (error.code) {
    case 'NETWORK_ERROR': return 'No Connection';
    case 'UNAUTHORIZED': return 'Session Expired';
    case 'NOT_FOUND': return 'Not Found';
    case 'FORBIDDEN': return 'Access Denied';
    default: return 'Something Went Wrong';
  }
}
```

---

## Error Handling Decision Matrix

| Situation | Go Backend | React Native Frontend |
|-----------|-----------|----------------------|
| Expected failure | Return `error` / `*AppError` | try/catch + `ApiError` |
| Input validation | `Validation()` AppError | Form validation (Zod) |
| Not found | `NotFound()` AppError | Show `EmptyState` |
| Network failure | Retry with backoff | Show retry button |
| Auth expired | `Unauthorized()` AppError | Redirect to login |
| Permission denied | `Forbidden()` AppError | Show access denied |
| WS disconnect | Auto-reconnect | Show reconnecting banner |
| Document lock conflict | `Conflict()` AppError | Show alert dialog |

---

## Logging Strategy (Go)

### Using `log/slog` (structured logging)

```go
import "log/slog"

func (s *ChatService) SendMessage(ctx context.Context, msg *Message) error {
    slog.Info("sending message",
        "chat_id", msg.ChatID,
        "sender_id", msg.SenderID,
    )

    if err := s.repo.Save(ctx, msg); err != nil {
        slog.Error("failed to save message",
            "chat_id", msg.ChatID,
            "error", err,
        )
        return err
    }

    slog.Debug("message saved successfully",
        "message_id", msg.ID,
    )
    return nil
}
```

### Log Levels

| Level | Usage | Example |
|-------|-------|---------|
| `Error` | Operasi gagal, perlu perhatian | Database down, push notif gagal |
| `Warn` | Tidak ideal tapi bisa lanjut | Retry diperlukan, cache miss |
| `Info` | Milestone penting | User login, message sent, document locked |
| `Debug` | Detail untuk debugging | WS message received, query executed |

---

## Forbidden Practices

| Practice | Why | Alternative |
|----------|-----|------------|
| `panic()` di handler | App crash | Return `error` |
| Ignore error (`_ = fn()`) | Bug tersembunyi | Handle atau log |
| `log.Fatal()` di handler | App exit | Log + return error |
| Generic error string | Tidak terstruktur | `AppError` types |
| Expose internal error ke user | Security risk | Sanitize di handler |
| Catch-all `catch (e) {}` | Error hilang | Log dan handle |
| `console.error` saja di RN | Tidak actionable | Show UI error + log |
