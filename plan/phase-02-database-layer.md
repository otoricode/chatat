# Phase 02: Database Layer

> Setup PostgreSQL schema, migrations, models, dan repositories.
> Phase ini menghasilkan data layer yang fully tested dan siap digunakan.

**Estimasi:** 3 hari
**Dependency:** Phase 01 (Project Setup)
**Output:** Database schema lengkap, migration system, models, dan repositories.

---

## Task 2.1: PostgreSQL Connection & Migration System

**Input:** Docker PostgreSQL dari Phase 01
**Output:** Database connection pool dan migration runner

### Steps:
1. Setup connection pool di `internal/database/postgres.go`:
   ```go
   func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
       config, err := pgxpool.ParseConfig(databaseURL)
       config.MaxConns = 25
       config.MinConns = 5
       config.MaxConnIdleTime = 5 * time.Minute
       return pgxpool.NewWithConfig(ctx, config)
   }
   ```
2. Install golang-migrate:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```
3. Tambah dependency:
   ```bash
   go get github.com/golang-migrate/migrate/v4
   go get github.com/golang-migrate/migrate/v4/database/postgres
   go get github.com/golang-migrate/migrate/v4/source/file
   ```
4. Buat `internal/database/migrate.go`:
   - Function `RunMigrations(databaseURL, migrationsPath string) error`
   - Auto-run saat server startup
5. Buat folder `migrations/` untuk SQL files
6. Tambah ke `Makefile`:
   ```makefile
   migrate-up:
   	migrate -path migrations -database "$(DATABASE_URL)" up

   migrate-down:
   	migrate -path migrations -database "$(DATABASE_URL)" down 1

   migrate-create:
   	migrate create -ext sql -dir migrations -seq $(name)
   ```
7. Test koneksi dan migration runner

### Acceptance Criteria:
- [x] Connection pool terbuat dan berfungsi
- [x] Migration runner auto-run saat startup
- [x] `make migrate-up` dan `make migrate-down` bekerja
- [x] `make migrate-create name=xxx` membuat file baru
- [x] Concurrent query tidak deadlock

### Testing:
- [x] Unit test: pool creation
- [x] Unit test: migration runner (up & down)
- [x] Unit test: concurrent queries

---

## Task 2.2: Redis Setup

**Input:** Docker Redis dari Phase 01
**Output:** Redis client yang berfungsi

### Steps:
1. Buat `internal/database/redis.go`:
   ```go
   func NewRedisClient(redisURL string) (*redis.Client, error) {
       opt, err := redis.ParseURL(redisURL)
       if err != nil {
           return nil, err
       }
       client := redis.NewClient(opt)
       // Ping test
       ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
       defer cancel()
       if err := client.Ping(ctx).Err(); err != nil {
           return nil, err
       }
       return client, nil
   }
   ```
2. Redis akan digunakan untuk:
   - Session storage (JWT blacklist)
   - OTP storage (temporary, TTL 5 menit)
   - Online status tracking
   - Typing indicator pub/sub
   - Rate limiting counters
3. Test connection dari main.go

### Acceptance Criteria:
- [x] Redis client terbuat dan berfungsi
- [x] Ping test sukses
- [x] Set/Get/Delete operations bekerja
- [x] TTL-based expiry bekerja

### Testing:
- [x] Unit test: connection
- [x] Unit test: set/get with TTL
- [x] Unit test: pub/sub basic

---

## Task 2.3: Schema — Users Table

**Input:** Task 2.1 selesai
**Output:** Users table, model, dan repository

### Steps:
1. Buat migration `000001_users.up.sql`:
   ```sql
   CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

   CREATE TABLE users (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     phone VARCHAR(20) NOT NULL UNIQUE,
     name VARCHAR(100) NOT NULL,
     avatar VARCHAR(10) DEFAULT '👤',
     status VARCHAR(200) DEFAULT '',
     last_seen TIMESTAMPTZ DEFAULT NOW(),
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_users_phone ON users(phone);
   CREATE INDEX idx_users_last_seen ON users(last_seen);
   ```
2. Buat `000001_users.down.sql`:
   ```sql
   DROP TABLE IF EXISTS users;
   ```
3. Buat `internal/model/user.go`:
   ```go
   type User struct {
       ID        uuid.UUID  `json:"id" db:"id"`
       Phone     string     `json:"phone" db:"phone"`
       Name      string     `json:"name" db:"name"`
       Avatar    string     `json:"avatar" db:"avatar"`
       Status    string     `json:"status" db:"status"`
       LastSeen  time.Time  `json:"lastSeen" db:"last_seen"`
       CreatedAt time.Time  `json:"createdAt" db:"created_at"`
       UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
   }

   type CreateUserInput struct {
       Phone  string `json:"phone" validate:"required,e164"`
       Name   string `json:"name" validate:"required,min=1,max=100"`
       Avatar string `json:"avatar"`
   }

   type UpdateUserInput struct {
       Name   *string `json:"name"`
       Avatar *string `json:"avatar"`
       Status *string `json:"status"`
   }
   ```
4. Buat `internal/repository/user_repo.go`:
   - `Create(ctx, input) (*User, error)`
   - `FindByID(ctx, id) (*User, error)`
   - `FindByPhone(ctx, phone) (*User, error)`
   - `FindByPhones(ctx, phones []string) ([]*User, error)` — untuk contact matching
   - `Update(ctx, id, input) (*User, error)`
   - `UpdateLastSeen(ctx, id) error`
   - `Delete(ctx, id) error`

### Acceptance Criteria:
- [x] Migration membuat tabel `users`
- [x] CRUD operations berfungsi
- [x] Phone number UNIQUE constraint enforced
- [x] UUID auto-generated
- [x] `FindByPhones` batch query berfungsi untuk contact matching
- [x] Index pada phone dan last_seen

### Testing:
- [x] Unit test: create user
- [x] Unit test: find by id (exists dan not found)
- [x] Unit test: find by phone
- [x] Unit test: find by phones (batch)
- [x] Unit test: update user
- [x] Unit test: delete user
- [x] Unit test: duplicate phone handling

---

## Task 2.4: Schema — Chats & Messages Tables

**Input:** Task 2.3 selesai
**Output:** Chat dan message tables, models, repositories

### Steps:
1. Buat migration `000002_chats.up.sql`:
   ```sql
   CREATE TABLE chats (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     type VARCHAR(10) NOT NULL CHECK(type IN ('personal', 'group')),
     name VARCHAR(100),
     icon VARCHAR(10),
     description TEXT,
     created_by UUID NOT NULL REFERENCES users(id),
     pinned_at TIMESTAMPTZ,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE TABLE chat_members (
     chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     role VARCHAR(10) NOT NULL DEFAULT 'member'
       CHECK(role IN ('admin', 'member')),
     joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (chat_id, user_id)
   );

   CREATE INDEX idx_chat_members_user_id ON chat_members(user_id);

   CREATE TABLE messages (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
     sender_id UUID NOT NULL REFERENCES users(id),
     content TEXT NOT NULL,
     reply_to_id UUID REFERENCES messages(id) ON DELETE SET NULL,
     type VARCHAR(20) NOT NULL DEFAULT 'text'
       CHECK(type IN ('text', 'image', 'file', 'document_card', 'system')),
     metadata JSONB,
     is_deleted BOOLEAN NOT NULL DEFAULT false,
     deleted_for_all BOOLEAN NOT NULL DEFAULT false,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_messages_chat_id ON messages(chat_id);
   CREATE INDEX idx_messages_chat_id_created_at ON messages(chat_id, created_at DESC);
   CREATE INDEX idx_messages_sender_id ON messages(sender_id);
   ```
2. Buat migration down file
3. Buat models di `internal/model/`:
   - `chat.go`: `Chat`, `ChatMember`, `ChatType` enum, `MemberRole` enum
   - `message.go`: `Message`, `MessageType` enum, `CreateMessageInput`
4. Buat repositories:
   - `chat_repo.go`:
     - `Create(ctx, input) (*Chat, error)`
     - `FindByID(ctx, id) (*Chat, error)`
     - `FindPersonalChat(ctx, userID1, userID2) (*Chat, error)`
     - `ListByUser(ctx, userID) ([]*ChatWithLastMessage, error)`
     - `AddMember(ctx, chatID, userID, role) error`
     - `RemoveMember(ctx, chatID, userID) error`
     - `GetMembers(ctx, chatID) ([]*ChatMember, error)`
     - `Update(ctx, id, input) (*Chat, error)`
     - `Delete(ctx, id) error`
   - `message_repo.go`:
     - `Create(ctx, input) (*Message, error)`
     - `FindByID(ctx, id) (*Message, error)`
     - `ListByChat(ctx, chatID, cursor, limit) ([]*Message, error)` — cursor pagination
     - `MarkAsDeleted(ctx, id, forAll) error`
     - `Search(ctx, chatID, query) ([]*Message, error)`

### Acceptance Criteria:
- [x] Chat type check constraint berfungsi
- [x] Member role check constraint berfungsi
- [x] CASCADE delete: hapus chat → hapus members + messages
- [x] Personal chat: unique pair (user1, user2)
- [x] Cursor pagination pada messages berfungsi
- [x] Reply to message reference berfungsi

### Testing:
- [x] Unit test: create personal chat
- [x] Unit test: create group chat
- [x] Unit test: add/remove member
- [x] Unit test: send message
- [x] Unit test: reply to message
- [x] Unit test: cursor pagination (order by created_at DESC)
- [x] Unit test: cascade delete
- [x] Unit test: soft delete message

---

## Task 2.5: Schema — Topics Table

**Input:** Task 2.4 selesai
**Output:** Topics table, model, repository

### Steps:
1. Buat migration `000003_topics.up.sql`:
   ```sql
   CREATE TABLE topics (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     name VARCHAR(100) NOT NULL,
     icon VARCHAR(10) NOT NULL DEFAULT '💬',
     description TEXT,
     parent_type VARCHAR(10) NOT NULL CHECK(parent_type IN ('personal', 'group')),
     parent_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
     created_by UUID NOT NULL REFERENCES users(id),
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE TABLE topic_members (
     topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     role VARCHAR(10) NOT NULL DEFAULT 'member'
       CHECK(role IN ('admin', 'member')),
     joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (topic_id, user_id)
   );

   CREATE INDEX idx_topics_parent_id ON topics(parent_id);
   CREATE INDEX idx_topic_members_user_id ON topic_members(user_id);

   CREATE TABLE topic_messages (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
     sender_id UUID NOT NULL REFERENCES users(id),
     content TEXT NOT NULL,
     reply_to_id UUID REFERENCES topic_messages(id) ON DELETE SET NULL,
     type VARCHAR(20) NOT NULL DEFAULT 'text'
       CHECK(type IN ('text', 'image', 'file', 'document_card', 'system')),
     metadata JSONB,
     is_deleted BOOLEAN NOT NULL DEFAULT false,
     deleted_for_all BOOLEAN NOT NULL DEFAULT false,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_topic_messages_topic_id ON topic_messages(topic_id);
   CREATE INDEX idx_topic_messages_topic_id_created_at ON topic_messages(topic_id, created_at DESC);
   ```
2. Buat migration down file
3. Buat `internal/model/topic.go`:
   ```go
   type Topic struct {
       ID          uuid.UUID `json:"id"`
       Name        string    `json:"name"`
       Icon        string    `json:"icon"`
       Description string    `json:"description"`
       ParentType  string    `json:"parentType"`
       ParentID    uuid.UUID `json:"parentId"`
       CreatedBy   uuid.UUID `json:"createdBy"`
       CreatedAt   time.Time `json:"createdAt"`
       UpdatedAt   time.Time `json:"updatedAt"`
   }
   ```
4. Buat `internal/repository/topic_repo.go`:
   - `Create(ctx, input) (*Topic, error)`
   - `FindByID(ctx, id) (*Topic, error)`
   - `ListByParent(ctx, parentID) ([]*Topic, error)`
   - `ListByUser(ctx, userID) ([]*Topic, error)`
   - `AddMember(ctx, topicID, userID, role) error`
   - `RemoveMember(ctx, topicID, userID) error`
   - `GetMembers(ctx, topicID) ([]*TopicMember, error)`
   - `Update(ctx, id, input) (*Topic, error)`
   - `Delete(ctx, id) error`
5. Buat `internal/repository/topic_message_repo.go` (same pattern as message_repo)

### Acceptance Criteria:
- [x] Topic parent must be a valid chat
- [x] Topic members must be members of parent chat
- [x] CASCADE: delete chat → delete topics → delete topic_messages
- [x] Topic messages support same features as chat messages

### Testing:
- [x] Unit test: create topic from personal chat
- [x] Unit test: create topic from group chat
- [x] Unit test: topic member validation (must be parent member)
- [x] Unit test: topic message CRUD
- [x] Unit test: cascade delete chain

---

## Task 2.6: Schema — Documents & Blocks Tables

**Input:** Task 2.4 selesai
**Output:** Document dan block tables, models, repositories

### Steps:
1. Buat migration `000004_documents.up.sql`:
   ```sql
   CREATE TABLE documents (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     title VARCHAR(200) NOT NULL,
     icon VARCHAR(10) NOT NULL DEFAULT '📄',
     cover VARCHAR(50),
     owner_id UUID NOT NULL REFERENCES users(id),
     chat_id UUID REFERENCES chats(id) ON DELETE SET NULL,
     topic_id UUID REFERENCES topics(id) ON DELETE SET NULL,
     is_standalone BOOLEAN NOT NULL DEFAULT false,
     require_sigs BOOLEAN NOT NULL DEFAULT false,
     locked BOOLEAN NOT NULL DEFAULT false,
     locked_at TIMESTAMPTZ,
     locked_by VARCHAR(20) CHECK(locked_by IN ('manual', 'signatures')),
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE TABLE document_collaborators (
     document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     role VARCHAR(10) NOT NULL DEFAULT 'editor'
       CHECK(role IN ('editor', 'viewer')),
     added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (document_id, user_id)
   );

   CREATE TABLE document_signers (
     document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     signed_at TIMESTAMPTZ,
     signer_name VARCHAR(100),
     PRIMARY KEY (document_id, user_id)
   );

   CREATE TABLE document_tags (
     document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
     tag VARCHAR(100) NOT NULL,
     PRIMARY KEY (document_id, tag)
   );

   CREATE TABLE blocks (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
     type VARCHAR(30) NOT NULL
       CHECK(type IN (
         'paragraph', 'heading1', 'heading2', 'heading3',
         'bullet-list', 'numbered-list', 'checklist',
         'table', 'callout', 'code', 'toggle', 'divider', 'quote'
       )),
     content TEXT,
     checked BOOLEAN,
     rows JSONB,
     columns JSONB,
     language VARCHAR(30),
     emoji VARCHAR(10),
     color VARCHAR(20),
     sort_order INTEGER NOT NULL DEFAULT 0,
     parent_block_id UUID REFERENCES blocks(id) ON DELETE CASCADE,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE TABLE document_history (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id),
     action VARCHAR(50) NOT NULL,
     details JSONB,
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE INDEX idx_documents_owner_id ON documents(owner_id);
   CREATE INDEX idx_documents_chat_id ON documents(chat_id);
   CREATE INDEX idx_documents_topic_id ON documents(topic_id);
   CREATE INDEX idx_documents_locked ON documents(locked);
   CREATE INDEX idx_blocks_document_id ON blocks(document_id);
   CREATE INDEX idx_blocks_sort_order ON blocks(document_id, sort_order);
   CREATE INDEX idx_document_history_document_id ON document_history(document_id);
   CREATE INDEX idx_document_tags_tag ON document_tags(tag);
   ```
2. Buat migration down file
3. Buat models:
   - `internal/model/document.go`: `Document`, `DocumentCollaborator`, `DocumentSigner`, `CollaboratorRole`
   - `internal/model/block.go`: `Block`, `BlockType` enum, `CreateBlockInput`, `UpdateBlockInput`
4. Buat repositories:
   - `document_repo.go`:
     - `Create(ctx, input) (*Document, error)`
     - `FindByID(ctx, id) (*Document, error)`
     - `ListByChat(ctx, chatID) ([]*Document, error)`
     - `ListByTopic(ctx, topicID) ([]*Document, error)`
     - `ListByOwner(ctx, ownerID) ([]*Document, error)`
     - `ListByTag(ctx, tag) ([]*Document, error)`
     - `AddCollaborator(ctx, docID, userID, role) error`
     - `RemoveCollaborator(ctx, docID, userID) error`
     - `AddSigner(ctx, docID, userID) error`
     - `RecordSignature(ctx, docID, userID, name) error`
     - `Lock(ctx, docID, lockedBy) error`
     - `AddTag(ctx, docID, tag) error`
     - `RemoveTag(ctx, docID, tag) error`
     - `Update(ctx, id, input) (*Document, error)`
     - `Delete(ctx, id) error`
   - `block_repo.go`:
     - `Create(ctx, input) (*Block, error)`
     - `ListByDocument(ctx, docID) ([]*Block, error)`
     - `Update(ctx, id, input) (*Block, error)`
     - `Reorder(ctx, docID, blockIDs []uuid.UUID) error`
     - `Delete(ctx, id) error`
   - `document_history_repo.go`:
     - `Create(ctx, input) error`
     - `ListByDocument(ctx, docID) ([]*DocumentHistory, error)`

### Acceptance Criteria:
- [x] Document context: chat_id XOR topic_id XOR standalone
- [x] Collaborator roles enforced
- [x] Signer tracking with timestamp
- [x] Block sort_order untuk ordering
- [x] Block parent (untuk toggle children)
- [x] Locked document tidak bisa di-update (enforce di service layer)
- [x] Tags many-to-many berfungsi
- [x] CASCADE delete: document → blocks, collaborators, signers, tags, history

### Testing:
- [x] Unit test: create document (in chat, in topic, standalone)
- [x] Unit test: collaborator CRUD
- [x] Unit test: signer flow (add, sign, check all signed)
- [x] Unit test: lock document (manual + signatures)
- [x] Unit test: block CRUD + reorder
- [x] Unit test: tag CRUD
- [x] Unit test: history logging
- [x] Unit test: cascade delete

---

## Task 2.7: Schema — Entities Table

**Input:** Task 2.6 selesai
**Output:** Entity tables, model, repository

### Steps:
1. Buat migration `000005_entities.up.sql`:
   ```sql
   CREATE TABLE entities (
     id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
     name VARCHAR(100) NOT NULL,
     type VARCHAR(50),
     owner_id UUID NOT NULL REFERENCES users(id),
     contact_user_id UUID REFERENCES users(id),
     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
   );

   CREATE TABLE document_entities (
     document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
     entity_id UUID NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
     PRIMARY KEY (document_id, entity_id)
   );

   CREATE INDEX idx_entities_owner_id ON entities(owner_id);
   CREATE INDEX idx_entities_contact_user_id ON entities(contact_user_id);
   CREATE INDEX idx_entities_name ON entities(name);
   CREATE INDEX idx_document_entities_entity_id ON document_entities(entity_id);
   ```
2. Buat migration down file
3. Buat `internal/model/entity.go`:
   ```go
   type Entity struct {
       ID            uuid.UUID  `json:"id"`
       Name          string     `json:"name"`
       Type          string     `json:"type"` // "lahan", "kendaraan", "kontak", etc.
       OwnerID       uuid.UUID  `json:"ownerId"`
       ContactUserID *uuid.UUID `json:"contactUserId"` // nil if not a contact entity
       CreatedAt     time.Time  `json:"createdAt"`
   }
   ```
4. Buat `internal/repository/entity_repo.go`:
   - `Create(ctx, input) (*Entity, error)`
   - `FindByID(ctx, id) (*Entity, error)`
   - `ListByOwner(ctx, ownerID) ([]*Entity, error)`
   - `Search(ctx, ownerID, query) ([]*Entity, error)`
   - `LinkToDocument(ctx, docID, entityID) error`
   - `UnlinkFromDocument(ctx, docID, entityID) error`
   - `ListByDocument(ctx, docID) ([]*Entity, error)`
   - `ListDocumentsByEntity(ctx, entityID) ([]*Document, error)`
   - `Delete(ctx, id) error`

### Acceptance Criteria:
- [x] Entity bisa free-form (nama + type bebas)
- [x] Contact entity link ke user
- [x] Many-to-many: document ↔ entity
- [x] Entity global per user (bisa dipakai di dokumen mana saja)
- [x] Search entity by name

### Testing:
- [x] Unit test: create entity (regular + contact)
- [x] Unit test: link/unlink entity to document
- [x] Unit test: list entities by document
- [x] Unit test: list documents by entity
- [x] Unit test: search entity
- [x] Unit test: cascade delete

---

## Task 2.8: Schema — Message Status & Read Receipts

**Input:** Task 2.4 selesai
**Output:** Tables untuk tracking delivery dan read status

### Steps:
1. Buat migration `000006_message_status.up.sql`:
   ```sql
   CREATE TABLE message_status (
     message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     status VARCHAR(10) NOT NULL DEFAULT 'sent'
       CHECK(status IN ('sent', 'delivered', 'read')),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (message_id, user_id)
   );

   CREATE INDEX idx_message_status_user_id ON message_status(user_id);

   CREATE TABLE topic_message_status (
     message_id UUID NOT NULL REFERENCES topic_messages(id) ON DELETE CASCADE,
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     status VARCHAR(10) NOT NULL DEFAULT 'sent'
       CHECK(status IN ('sent', 'delivered', 'read')),
     updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (message_id, user_id)
   );
   ```
2. Buat migration down file
3. Buat `internal/model/message_status.go`
4. Buat `internal/repository/message_status_repo.go`:
   - `Create(ctx, messageID, userID, status) error`
   - `UpdateStatus(ctx, messageID, userID, status) error`
   - `GetStatus(ctx, messageID) ([]*MessageStatus, error)`
   - `MarkChatAsRead(ctx, chatID, userID) error` — batch update
   - `GetUnreadCount(ctx, chatID, userID) (int, error)`

### Acceptance Criteria:
- [x] Status: sent → delivered → read (one-way)
- [x] Per-user tracking untuk group messages
- [x] Batch mark-as-read berfungsi
- [x] Unread count query efisien

### Testing:
- [x] Unit test: status transitions
- [x] Unit test: batch mark as read
- [x] Unit test: unread count
- [x] Unit test: cascade delete

---

## Phase 02 Review

### Testing Checklist:
- [x] Semua migration up tanpa error
- [x] Semua migration down tanpa error (rollback)
- [x] Semua repository CRUD tested
- [x] Cascade delete tested end-to-end
- [x] Constraint violations handled gracefully
- [x] Concurrent access tested
- [x] `go test ./...` — semua test pass

### Review Checklist:
- [x] Schema sesuai `spesifikasi-chatat.md` data structures
- [x] Model structs sesuai `docs/go-style-guide.md`
- [x] Error handling sesuai `docs/error-handling.md`
- [x] Naming sesuai `docs/naming-conventions.md`
- [x] Migration files sequential dan atomic
- [x] No panic/fatal di repository code
- [x] Commit: `feat(db): implement database layer with migrations`
