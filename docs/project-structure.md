# Project Structure

> Definisi struktur folder dan file untuk Chatat

---

## Root Structure

```
chatat/
├── server/                     # Go backend (API + WebSocket)
├── mobile/                     # React Native frontend
├── proto/                      # Protocol Buffers / API contracts
├── migrations/                 # Database migrations
├── scripts/                    # Build & dev scripts
├── docs/                       # Development documentation
├── .github/                    # CI/CD workflows
└── README.md
```

---

## Backend: `server/`

```
server/
├── go.mod                      # Go module definition
├── go.sum                      # Dependency checksums
├── main.go                     # Entry point
├── Dockerfile                  # Container build
│
├── cmd/                        # CLI entry points
│   └── server/
│       └── main.go             # Server bootstrap
│
├── internal/                   # Private application code
│   ├── config/                 # Configuration
│   │   ├── config.go           # Config struct + loading
│   │   └── env.go              # Environment variable mapping
│   │
│   ├── auth/                   # Authentication
│   │   ├── handler.go          # HTTP handlers (OTP, verify)
│   │   ├── service.go          # Auth business logic
│   │   ├── otp.go              # OTP generation + validation
│   │   ├── reverse_otp.go      # Reverse OTP via WhatsApp
│   │   └── token.go            # JWT token management
│   │
│   ├── user/                   # User domain
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # User business logic
│   │   ├── repository.go       # Database operations
│   │   └── model.go            # User structs
│   │
│   ├── chat/                   # Chat domain
│   │   ├── handler.go          # HTTP + WebSocket handlers
│   │   ├── service.go          # Chat business logic
│   │   ├── repository.go       # Database operations
│   │   ├── model.go            # Chat, Message structs
│   │   └── hub.go              # WebSocket connection hub
│   │
│   ├── group/                  # Group domain
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # Group business logic
│   │   ├── repository.go       # Database operations
│   │   └── model.go            # Group structs
│   │
│   ├── topic/                  # Topic domain
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # Topic business logic
│   │   ├── repository.go       # Database operations
│   │   └── model.go            # Topic structs
│   │
│   ├── document/               # Document domain
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # Document business logic
│   │   ├── repository.go       # Database operations
│   │   ├── model.go            # Document, Block structs
│   │   ├── lock.go             # Document locking logic
│   │   └── collab.go           # Real-time collaboration (CRDT/OT)
│   │
│   ├── entity/                 # Entity/tag domain
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # Entity business logic
│   │   ├── repository.go       # Database operations
│   │   └── model.go            # Entity structs
│   │
│   ├── contact/                # Contact domain
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # Contact matching logic
│   │   ├── repository.go       # Database operations
│   │   └── model.go            # Contact structs
│   │
│   ├── media/                  # Media upload/download
│   │   ├── handler.go          # HTTP handlers
│   │   ├── service.go          # Media processing
│   │   ├── storage.go          # File storage (S3/local)
│   │   └── model.go            # Media structs
│   │
│   ├── notification/           # Push notifications
│   │   ├── service.go          # Notification logic
│   │   ├── fcm.go              # Firebase Cloud Messaging
│   │   └── apns.go             # Apple Push Notification
│   │
│   ├── ws/                     # WebSocket infrastructure
│   │   ├── hub.go              # Connection manager
│   │   ├── client.go           # Single client connection
│   │   ├── message.go          # WS message types
│   │   └── handler.go          # WS event dispatcher
│   │
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go             # JWT authentication
│   │   ├── cors.go             # CORS configuration
│   │   ├── ratelimit.go        # Rate limiting
│   │   ├── logging.go          # Request logging
│   │   └── recovery.go         # Panic recovery
│   │
│   └── database/               # Database infrastructure
│       ├── postgres.go         # PostgreSQL connection
│       ├── redis.go            # Redis connection (cache/pubsub)
│       └── migration.go        # Migration runner
│
├── pkg/                        # Shared/public packages
│   ├── validator/              # Input validation
│   │   └── validator.go
│   ├── crypto/                 # Encryption utilities
│   │   ├── e2e.go              # End-to-end encryption
│   │   └── hash.go             # Hashing utilities
│   ├── apperror/               # Application error types
│   │   └── errors.go
│   └── response/               # HTTP response helpers
│       └── json.go
│
└── test/                       # Integration tests
    ├── fixtures/               # Test data
    ├── helpers.go              # Test utilities
    ├── auth_test.go
    ├── chat_test.go
    ├── document_test.go
    └── topic_test.go
```

---

## Frontend: `mobile/`

```
mobile/
├── package.json                # Dependencies
├── tsconfig.json               # TypeScript config
├── app.json                    # Expo/RN config
├── babel.config.js             # Babel config
├── metro.config.js             # Metro bundler config
├── index.js                    # Entry point
│
├── src/
│   ├── App.tsx                 # Root component + navigation
│   │
│   ├── screens/                # Screen components (pages)
│   │   ├── auth/
│   │   │   ├── LoginScreen.tsx         # Phone number input
│   │   │   ├── OTPScreen.tsx           # OTP verification
│   │   │   └── ProfileSetupScreen.tsx  # Initial profile setup
│   │   │
│   │   ├── chat/
│   │   │   ├── ChatListScreen.tsx      # Chat list (home)
│   │   │   ├── ChatScreen.tsx          # Chat conversation
│   │   │   ├── ChatInfoScreen.tsx      # Chat/Group info
│   │   │   └── NewChatScreen.tsx       # Start new chat
│   │   │
│   │   ├── group/
│   │   │   ├── NewGroupScreen.tsx      # Create group
│   │   │   ├── GroupInfoScreen.tsx     # Group details
│   │   │   └── GroupMembersScreen.tsx  # Manage members
│   │   │
│   │   ├── topic/
│   │   │   ├── TopicListScreen.tsx     # Topics in chat/group
│   │   │   ├── TopicScreen.tsx         # Topic conversation
│   │   │   └── NewTopicScreen.tsx      # Create topic
│   │   │
│   │   ├── document/
│   │   │   ├── DocumentListScreen.tsx  # Documents tab
│   │   │   ├── DocumentScreen.tsx      # Document editor
│   │   │   ├── DocumentInfoScreen.tsx  # Document details
│   │   │   └── NewDocumentScreen.tsx   # Create document
│   │   │
│   │   ├── contact/
│   │   │   ├── ContactListScreen.tsx   # Contacts list
│   │   │   └── ContactInfoScreen.tsx   # Contact details
│   │   │
│   │   └── settings/
│   │       ├── SettingsScreen.tsx      # Settings menu
│   │       ├── ProfileScreen.tsx       # Edit profile
│   │       └── LanguageScreen.tsx      # Language selection
│   │
│   ├── components/             # Reusable UI components
│   │   ├── ui/                 # Primitives
│   │   │   ├── Button.tsx
│   │   │   ├── TextInput.tsx
│   │   │   ├── Avatar.tsx
│   │   │   ├── Badge.tsx
│   │   │   ├── Icon.tsx
│   │   │   ├── Modal.tsx
│   │   │   ├── Toast.tsx
│   │   │   └── Skeleton.tsx
│   │   │
│   │   ├── chat/               # Chat-specific components
│   │   │   ├── MessageBubble.tsx
│   │   │   ├── MessageInput.tsx
│   │   │   ├── ChatListItem.tsx
│   │   │   ├── TypingIndicator.tsx
│   │   │   ├── ReadReceipt.tsx
│   │   │   └── DocumentCard.tsx    # Inline document card
│   │   │
│   │   ├── document/           # Document-specific components
│   │   │   ├── BlockRenderer.tsx
│   │   │   ├── BlockEditor.tsx
│   │   │   ├── BlockToolbar.tsx
│   │   │   ├── TableBlock.tsx
│   │   │   ├── ChecklistBlock.tsx
│   │   │   ├── CalloutBlock.tsx
│   │   │   ├── CodeBlock.tsx
│   │   │   ├── LockBanner.tsx
│   │   │   └── SignatureSheet.tsx
│   │   │
│   │   ├── entity/             # Entity components
│   │   │   ├── EntityPicker.tsx
│   │   │   └── EntityBadge.tsx
│   │   │
│   │   └── shared/             # Shared components
│   │       ├── EmptyState.tsx
│   │       ├── LoadingState.tsx
│   │       ├── ErrorState.tsx
│   │       ├── SearchBar.tsx
│   │       ├── ConfirmDialog.tsx
│   │       └── FloatingButton.tsx
│   │
│   ├── hooks/                  # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useChat.ts
│   │   ├── useTopic.ts
│   │   ├── useDocument.ts
│   │   ├── useEntity.ts
│   │   ├── useContacts.ts
│   │   ├── useWebSocket.ts
│   │   ├── useNotification.ts
│   │   └── useStorage.ts          # Local SQLite storage
│   │
│   ├── stores/                 # Zustand state stores
│   │   ├── authStore.ts
│   │   ├── chatStore.ts
│   │   ├── topicStore.ts
│   │   ├── documentStore.ts
│   │   ├── contactStore.ts
│   │   └── uiStore.ts
│   │
│   ├── services/               # API & network layer
│   │   ├── api.ts              # Base HTTP client (axios/fetch)
│   │   ├── authApi.ts          # Auth endpoints
│   │   ├── chatApi.ts          # Chat endpoints
│   │   ├── topicApi.ts         # Topic endpoints
│   │   ├── documentApi.ts      # Document endpoints
│   │   ├── contactApi.ts       # Contact endpoints
│   │   ├── mediaApi.ts         # Media upload/download
│   │   └── wsClient.ts        # WebSocket client
│   │
│   ├── database/               # Local database (SQLite)
│   │   ├── schema.ts           # Table definitions
│   │   ├── migrations.ts       # Local migrations
│   │   ├── chatDb.ts           # Chat local queries
│   │   ├── documentDb.ts       # Document local queries
│   │   └── syncEngine.ts       # Sync local <-> server
│   │
│   ├── lib/                    # Utility libraries
│   │   ├── constants.ts        # App constants
│   │   ├── formatting.ts       # Date, number formatting
│   │   ├── validation.ts       # Input validation
│   │   ├── encryption.ts       # E2E encryption utils
│   │   └── i18n.ts             # Internationalization (ID/EN/AR)
│   │
│   ├── types/                  # TypeScript type definitions
│   │   ├── user.ts
│   │   ├── chat.ts
│   │   ├── topic.ts
│   │   ├── document.ts
│   │   ├── entity.ts
│   │   ├── contact.ts
│   │   ├── message.ts
│   │   └── api.ts              # API request/response types
│   │
│   ├── navigation/             # React Navigation setup
│   │   ├── RootNavigator.tsx
│   │   ├── AuthNavigator.tsx
│   │   ├── MainNavigator.tsx
│   │   └── types.ts            # Navigation param types
│   │
│   ├── i18n/                   # Translation files
│   │   ├── id.json             # Bahasa Indonesia
│   │   ├── en.json             # English
│   │   └── ar.json             # Arabic
│   │
│   └── assets/                 # Static assets
│       ├── images/
│       └── fonts/
│
└── __tests__/                  # Test files
    ├── screens/
    ├── components/
    ├── hooks/
    └── stores/
```

---

## Database Migrations: `migrations/`

```
migrations/
├── 001_create_users.sql
├── 002_create_chats.sql
├── 003_create_messages.sql
├── 004_create_groups.sql
├── 005_create_topics.sql
├── 006_create_documents.sql
├── 007_create_blocks.sql
├── 008_create_entities.sql
├── 009_create_media.sql
└── 010_create_notifications.sql
```

---

## File Naming Rules

| Context | Convention | Example |
|---------|-----------|---------|
| Go source | snake_case | `reverse_otp.go` |
| Go test | `*_test.go` | `service_test.go` |
| Go package | lowercase, no underscore | `package auth` |
| React Native screens | PascalCase + Screen suffix | `ChatScreen.tsx` |
| React Native components | PascalCase | `MessageBubble.tsx` |
| React hooks | camelCase with `use` prefix | `useChat.ts` |
| Zustand stores | camelCase with `Store` suffix | `chatStore.ts` |
| TypeScript types | camelCase filename, PascalCase exports | `chat.ts` → `ChatMessage` |
| API services | camelCase with `Api` suffix | `chatApi.ts` |
| Local DB modules | camelCase with `Db` suffix | `chatDb.ts` |
| Translation files | lowercase ISO 639-1 | `id.json`, `en.json`, `ar.json` |
| SQL migrations | numbered prefix | `001_create_users.sql` |
| Test files | co-located or `__tests__/` | `ChatScreen.test.tsx` |

---

## Import Organization

### Go

```go
import (
    // 1. Standard library
    "context"
    "fmt"
    "net/http"

    // 2. External packages
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"

    // 3. Internal packages
    "github.com/otoritech/chatat/internal/auth"
    "github.com/otoritech/chatat/pkg/apperror"
)
```

### TypeScript

```typescript
// 1. React / React Native
import React, { useState, useEffect } from 'react';
import { View, Text, FlatList } from 'react-native';

// 2. External libraries
import { useNavigation } from '@react-navigation/native';

// 3. Internal: components
import { MessageBubble } from '@/components/chat/MessageBubble';
import { Button } from '@/components/ui/Button';

// 4. Internal: hooks, stores, services
import { useChat } from '@/hooks/useChat';
import { chatStore } from '@/stores/chatStore';

// 5. Internal: types
import type { ChatMessage } from '@/types/message';
```
