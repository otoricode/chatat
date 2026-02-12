# Testing Strategy

> Strategi testing untuk backend (Go) dan frontend (React Native/TypeScript).
> Fokus pada reliability dan confidence tanpa over-testing.

---

## Testing Pyramid

```
         /  E2E  \           ← Sedikit, critical paths saja
        /----------\
       / Integration \       ← API endpoints, database, WebSocket
      /----------------\
     /    Unit Tests     \   ← Banyak, cepat, isolated
    /______________________\
```

| Layer | Tool | Coverage Target | Jumlah |
|-------|------|----------------|--------|
| Unit (Go) | `go test` | 80%+ | Banyak |
| Unit (TS) | Jest | 70%+ | Banyak |
| Integration (Go) | `go test` + test DB | 60%+ | Moderate |
| Component (RN) | React Native Testing Library | 50%+ | Moderate |
| E2E | Detox | Critical paths | Sedikit |

---

## Backend Testing (Go)

### Unit Test Pattern

```go
// internal/chat/service_test.go

package chat

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestChatService_SendMessage(t *testing.T) {
    t.Run("success - sends message to valid chat", func(t *testing.T) {
        repo := new(MockChatRepository)
        wsHub := new(MockHub)
        notifier := new(MockNotifier)

        chat := &Chat{
            ID:        "chat-1",
            Type:      ChatTypePersonal,
            MemberIDs: []string{"user-1", "user-2"},
        }

        repo.On("FindByID", mock.Anything, "chat-1").Return(chat, nil)
        repo.On("SaveMessage", mock.Anything, "chat-1", mock.AnythingOfType("*Message")).Return(nil)
        wsHub.On("BroadcastToChat", mock.Anything, mock.Anything, mock.Anything).Return()
        notifier.On("NotifyChatMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

        svc := NewChatService(repo, nil, wsHub, notifier)

        msg, err := svc.SendMessage(context.Background(), "user-1", "chat-1", SendMessageRequest{
            Text: "Hello",
        })

        assert.NoError(t, err)
        assert.Equal(t, "Hello", msg.Text)
        assert.Equal(t, "user-1", msg.SenderID)
        repo.AssertExpectations(t)
    })

    t.Run("error - user is not a member", func(t *testing.T) {
        repo := new(MockChatRepository)

        chat := &Chat{
            ID:        "chat-1",
            MemberIDs: []string{"user-1", "user-2"},
        }

        repo.On("FindByID", mock.Anything, "chat-1").Return(chat, nil)

        svc := NewChatService(repo, nil, nil, nil)

        _, err := svc.SendMessage(context.Background(), "user-3", "chat-1", SendMessageRequest{
            Text: "Hello",
        })

        assert.Error(t, err)
        assert.Contains(t, err.Error(), "forbidden")
    })

    t.Run("error - chat not found", func(t *testing.T) {
        repo := new(MockChatRepository)
        repo.On("FindByID", mock.Anything, "invalid").Return(nil, apperror.NotFound("chat", "invalid"))

        svc := NewChatService(repo, nil, nil, nil)

        _, err := svc.SendMessage(context.Background(), "user-1", "invalid", SendMessageRequest{
            Text: "Hello",
        })

        assert.Error(t, err)
    })
}
```

### Naming Convention

```go
func Test{Type}_{Method}(t *testing.T) {
    t.Run("{scenario}", func(t *testing.T) { ... })
}

// Examples:
func TestChatService_SendMessage(t *testing.T) { ... }
func TestDocumentService_Lock(t *testing.T) { ... }
func TestOTPService_Validate(t *testing.T) { ... }
```

### Table-Driven Tests

```go
func TestDocument_HasSignedAll(t *testing.T) {
    tests := []struct {
        name      string
        signerIDs []string
        sigs      map[string]Signature
        want      bool
    }{
        {
            name:      "all signed",
            signerIDs: []string{"user-1", "user-2"},
            sigs: map[string]Signature{
                "user-1": {At: time.Now()},
                "user-2": {At: time.Now()},
            },
            want: true,
        },
        {
            name:      "partially signed",
            signerIDs: []string{"user-1", "user-2"},
            sigs: map[string]Signature{
                "user-1": {At: time.Now()},
            },
            want: false,
        },
        {
            name:      "no signers required",
            signerIDs: []string{},
            sigs:      map[string]Signature{},
            want:      true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            doc := &Document{
                SignerIDs: tt.signerIDs,
                Sigs:      tt.sigs,
            }
            assert.Equal(t, tt.want, doc.HasSignedAll())
        })
    }
}
```

### Mock Interfaces

```go
// internal/chat/mock_test.go

type MockChatRepository struct {
    mock.Mock
}

func (m *MockChatRepository) FindByID(ctx context.Context, id string) (*Chat, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*Chat), args.Error(1)
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, chatID string, msg *Message) error {
    args := m.Called(ctx, chatID, msg)
    return args.Error(0)
}

func (m *MockChatRepository) Create(ctx context.Context, chat *Chat) error {
    args := m.Called(ctx, chat)
    return args.Error(0)
}
```

### Integration Tests (Database)

```go
// test/chat_test.go

package test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestChatRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    db := setupTestDB(t)
    defer db.Close()

    repo := chat.NewChatRepository(db)
    ctx := context.Background()

    t.Run("create and find chat", func(t *testing.T) {
        c := &chat.Chat{
            ID:        "test-chat-1",
            Type:      chat.ChatTypePersonal,
            MemberIDs: []string{"user-1", "user-2"},
        }

        err := repo.Create(ctx, c)
        require.NoError(t, err)

        found, err := repo.FindByID(ctx, "test-chat-1")
        require.NoError(t, err)
        assert.Equal(t, c.ID, found.ID)
        assert.Equal(t, c.Type, found.Type)
    })

    t.Run("find non-existent returns not found", func(t *testing.T) {
        _, err := repo.FindByID(ctx, "non-existent")
        assert.Error(t, err)
    })
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    // Create isolated test database
    pool, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
    require.NoError(t, err)
    runMigrations(t, pool)
    return pool
}
```

---

## Frontend Testing (React Native)

### Test Runner: Jest + React Native Testing Library

```typescript
// jest.config.js
module.exports = {
  preset: 'react-native',
  setupFilesAfterSetup: ['./src/test/setup.ts'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
  },
  collectCoverageFrom: [
    'src/**/*.{ts,tsx}',
    '!src/**/*.test.{ts,tsx}',
    '!src/types/**',
  ],
};
```

### Setup File

```typescript
// src/test/setup.ts
import '@testing-library/jest-native/extend-expect';

// Mock async storage
jest.mock('@react-native-async-storage/async-storage', () =>
  require('@react-native-async-storage/async-storage/jest/async-storage-mock')
);

// Mock navigation
jest.mock('@react-navigation/native', () => ({
  useNavigation: () => ({ navigate: jest.fn(), goBack: jest.fn() }),
  useRoute: () => ({ params: {} }),
}));
```

### Component Test

```typescript
// components/chat/MessageBubble.test.tsx
import { render, screen } from '@testing-library/react-native';
import { MessageBubble } from './MessageBubble';

describe('MessageBubble', () => {
  const message = {
    id: 'msg-1',
    text: 'Hello there',
    senderId: 'user-1',
    createdAt: '2025-01-01T12:00:00Z',
  };

  it('renders message text', () => {
    render(<MessageBubble message={message} isMine={false} />);
    expect(screen.getByText('Hello there')).toBeTruthy();
  });

  it('applies mine style when isMine is true', () => {
    render(<MessageBubble message={message} isMine={true} />);
    const bubble = screen.getByTestId('message-bubble');
    expect(bubble).toHaveStyle({ alignSelf: 'flex-end' });
  });

  it('shows timestamp', () => {
    render(<MessageBubble message={message} isMine={false} />);
    expect(screen.getByText('12:00')).toBeTruthy();
  });
});
```

### Hook Test

```typescript
// hooks/useDocument.test.ts
import { renderHook, waitFor, act } from '@testing-library/react-native';
import { useDocument } from './useDocument';
import { documentApi } from '@/services/documentApi';

jest.mock('@/services/documentApi');

describe('useDocument', () => {
  const mockDocument = {
    id: 'doc-1',
    title: 'Test Document',
    locked: false,
    blocks: [],
  };

  beforeEach(() => jest.clearAllMocks());

  it('fetches document by id', async () => {
    (documentApi.getById as jest.Mock).mockResolvedValue(mockDocument);

    const { result } = renderHook(() => useDocument('doc-1'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(documentApi.getById).toHaveBeenCalledWith('doc-1');
    expect(result.current.document).toEqual(mockDocument);
    expect(result.current.error).toBeNull();
  });

  it('handles error', async () => {
    (documentApi.getById as jest.Mock).mockRejectedValue(new Error('Not found'));

    const { result } = renderHook(() => useDocument('invalid'));

    await waitFor(() => {
      expect(result.current.isLoading).toBe(false);
    });

    expect(result.current.document).toBeNull();
    expect(result.current.error).toBeTruthy();
  });
});
```

### Store Test

```typescript
// stores/chatStore.test.ts
import { useChatStore } from './chatStore';

describe('chatStore', () => {
  beforeEach(() => {
    useChatStore.setState({
      chats: [],
      activeChatId: null,
      unreadCounts: {},
    });
  });

  it('increments unread count', () => {
    useChatStore.getState().incrementUnread('chat-1');
    useChatStore.getState().incrementUnread('chat-1');

    expect(useChatStore.getState().unreadCounts['chat-1']).toBe(2);
  });

  it('clears unread count', () => {
    useChatStore.setState({ unreadCounts: { 'chat-1': 5 } });
    useChatStore.getState().clearUnread('chat-1');

    expect(useChatStore.getState().unreadCounts['chat-1']).toBe(0);
  });
});
```

---

## What to Test vs What Not to Test

### WAJIB Test

| Item | Why |
|------|-----|
| Message delivery logic | Core feature, harus reliable |
| Document locking | Race conditions, harus presisi |
| OTP validation | Security critical |
| Permission checks | Authorization harus benar |
| WebSocket event routing | Harus sampai ke client yang tepat |
| Database CRUD | Data integrity |
| Input validation | User input unpredictable |
| E2E encryption | Security critical |

### TIDAK PERLU Test

| Item | Why |
|------|-----|
| React Native built-in components | Already tested by framework |
| Third-party library internals | Not our code |
| Simple pass-through components | No logic |
| Static/presentational screens | Just layout |
| Styling/layout | Visual, bukan logic |
| Navigation config | Framework handles it |

---

## Test Organization

### Backend

```
server/
├── internal/
│   ├── chat/
│   │   ├── service.go
│   │   ├── service_test.go          ← Unit tests (co-located)
│   │   ├── mock_test.go             ← Mocks (co-located)
│   │   └── repository.go
│   ├── document/
│   │   ├── service.go
│   │   ├── service_test.go
│   │   └── lock_test.go
│   └── auth/
│       ├── otp.go
│       └── otp_test.go
│
└── test/                             ← Integration tests
    ├── fixtures/
    │   └── testdata.go
    ├── helpers.go
    ├── chat_test.go
    ├── document_test.go
    └── auth_test.go
```

### Frontend

```
mobile/src/
├── components/
│   ├── chat/
│   │   ├── MessageBubble.tsx
│   │   └── MessageBubble.test.tsx    ← Co-located
│   └── document/
│       ├── BlockRenderer.tsx
│       └── BlockRenderer.test.tsx
├── hooks/
│   ├── useChat.ts
│   └── useChat.test.ts              ← Co-located
├── stores/
│   ├── chatStore.ts
│   └── chatStore.test.ts            ← Co-located
└── test/
    ├── setup.ts                     ← Global test setup
    └── helpers.ts                   ← Shared test utilities
```

---

## CI Integration

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test-go:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16
        env:
          POSTGRES_DB: chatat_test
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports: ['5432:5432']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: cd server && go test ./...
      - run: cd server && go vet ./...
      - run: cd server && golangci-lint run

  test-mobile:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: 20 }
      - run: cd mobile && npm ci
      - run: cd mobile && npm run lint
      - run: cd mobile && npm test -- --coverage
      - run: cd mobile && npx tsc --noEmit
```

---

## Test Commands

```bash
# Go
cd server
go test ./...                           # All tests
go test ./internal/chat/...             # Package tests
go test -short ./...                    # Skip integration tests
go test -v ./internal/chat/...          # Verbose
go test -race ./...                     # Race condition detection
go test -cover ./...                    # Coverage report
go test -run TestChatService ./...      # Specific test

# React Native
cd mobile
npx jest                                # All tests
npx jest --watch                        # Watch mode
npx jest --coverage                     # With coverage
npx jest MessageBubble                  # Tests matching name
```
