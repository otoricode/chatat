# Phase 05: User & Contact System

> Implementasi profil pengguna dan sistem kontak berbasis nomor HP.
> Phase ini menghasilkan contact matching dan user management.

**Estimasi:** 3 hari
**Dependency:** Phase 04 (Authentication)
**Output:** User profile management dan contact sync berfungsi.

---

## Task 5.1: User Profile Service & Handler

**Input:** User model dan repository dari Phase 02
**Output:** User profile CRUD endpoints

### Steps:
1. Buat `internal/service/user_service.go`:
   ```go
   type UserService interface {
       GetProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
       UpdateProfile(ctx context.Context, userID uuid.UUID, input model.UpdateUserInput) (*model.User, error)
       SetupProfile(ctx context.Context, userID uuid.UUID, name string, avatar string) (*model.User, error)
       UpdateLastSeen(ctx context.Context, userID uuid.UUID) error
       DeleteAccount(ctx context.Context, userID uuid.UUID) error
   }
   ```
2. Buat `internal/handler/user_handler.go`:
   - `GET /api/v1/users/me` → get own profile
   - `PUT /api/v1/users/me` → update profile (name, avatar, status)
   - `POST /api/v1/users/me/setup` → first-time profile setup (setelah register)
     - Body: `{"name": "Andi", "avatar": "😊"}`
     - Validate: name required, avatar from emoji set
   - `DELETE /api/v1/users/me` → delete account (with confirmation)
3. Implementasi avatar:
   - Avatar = emoji string (single emoji character)
   - Predefined: 👤 😊 🙋 😎 🤓 👩 👨 🧑 dll
   - User bisa pick from emoji keyboard
4. Implementasi status:
   - Max 200 characters
   - Optional (bisa kosong)
5. LastSeen update:
   - Auto-update saat:
     - WebSocket pong received
     - API request authenticated
   - Debounce: max 1 update per 30 detik

### Acceptance Criteria:
- [x] Get own profile
- [x] Update name, avatar, status
- [x] First-time setup: name + avatar (required setelah register)
- [x] Delete account: cascade delete all user data
- [x] LastSeen auto-update dengan debounce
- [x] Avatar validation (valid emoji)

### Testing:
- [x] Unit test: get profile
- [x] Unit test: update profile (partial update)
- [x] Unit test: setup profile (validation)
- [x] Unit test: delete account
- [x] Unit test: last seen debounce

---

## Task 5.2: Contact Sync Service

**Input:** Task 5.1, User repository
**Output:** Contact matching berdasarkan nomor HP

### Steps:
1. Buat `internal/service/contact_service.go`:
   ```go
   type ContactService interface {
       SyncContacts(ctx context.Context, userID uuid.UUID, phoneHashes []string) ([]*ContactMatch, error)
       GetContacts(ctx context.Context, userID uuid.UUID) ([]*ContactInfo, error)
       SearchByPhone(ctx context.Context, phone string) (*model.User, error)
   }

   type ContactMatch struct {
       PhoneHash string    `json:"phoneHash"`
       UserID    uuid.UUID `json:"userId"`
       Name      string    `json:"name"`
       Avatar    string    `json:"avatar"`
       Status    string    `json:"status"`
       LastSeen  time.Time `json:"lastSeen"`
   }

   type ContactInfo struct {
       UserID    uuid.UUID `json:"userId"`
       Phone     string    `json:"phone"`
       Name      string    `json:"name"`
       Avatar    string    `json:"avatar"`
       Status    string    `json:"status"`
       IsOnline  bool      `json:"isOnline"`
       LastSeen  time.Time `json:"lastSeen"`
   }
   ```
2. Implementasi SyncContacts:
   - Mobile app mengirim SHA-256 hash dari semua nomor kontak
   - Server match hashes dengan hash nomor HP users terdaftar
   - Return: list matched users (tanpa expose phone numbers)
   - Flow:
     1. App normalize semua kontak → hash → kirim ke server
     2. Server: `SELECT * FROM users WHERE SHA256(phone) IN (...hashes)`
     3. Return matched users
3. Create contacts cache table:
   ```sql
   CREATE TABLE user_contacts (
     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     contact_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     contact_name VARCHAR(100),
     synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
     PRIMARY KEY (user_id, contact_user_id)
   );
   ```
   - Cache matched contacts for quick retrieval
   - Re-sync periodically (daily) atau saat user triggers manual sync
4. Online status:
   - Check WebSocket Hub: `hub.IsOnline(userID)`
   - Combine with last_seen timestamp

### Acceptance Criteria:
- [x] Phone hash matching berfungsi
- [x] Privacy: server never receives plain phone numbers dari kontak
- [x] Matched contacts returned dengan profile info
- [x] Contact cache untuk quick retrieval
- [x] Online status accurate (via WS hub)
- [x] Search by phone number (manual add contact)

### Testing:
- [x] Unit test: sync contacts (hash matching)
- [x] Unit test: no matches scenario
- [x] Unit test: partial matches
- [x] Unit test: get cached contacts
- [x] Unit test: search by phone
- [x] Unit test: online status integration

---

## Task 5.3: Contact Handler & Endpoints

**Input:** Task 5.2 selesai
**Output:** Contact REST endpoints

### Steps:
1. Buat `internal/handler/contact_handler.go`:
   - `POST /api/v1/contacts/sync`:
     - Body: `{"phoneHashes": ["abc123...", "def456..."]}`
     - Sync dan cache contacts
     - Response: `{"data": [{"phoneHash": "abc123", "userId": "...", "name": "Andi", ...}]}`
   - `GET /api/v1/contacts`:
     - Return cached contact list dengan online status
     - Sorted by name
     - Response: `{"data": [{"userId": "...", "phone": "+62...", "name": "Andi", "isOnline": true, ...}]}`
   - `GET /api/v1/contacts/search?phone=+628xxx`:
     - Search specific phone number
     - Return user if exists and registered
   - `GET /api/v1/contacts/:userId`:
     - Get specific contact profile
     - Return user info + online status
2. Buat migration `000007_user_contacts.up.sql` (contact cache table)
3. Contact list sorting:
   - Online users first
   - Then sorted alphabetically by name

### Acceptance Criteria:
- [x] Sync endpoint accepts phone hashes, returns matches
- [x] Contact list sorted: online first, then alphabetical
- [x] Search by phone number
- [x] Individual contact profile
- [x] Privacy preserved (hashed phone matching)

### Testing:
- [x] Unit test: full sync flow
- [x] Unit test: contact list retrieval
- [x] Unit test: search by phone
- [x] Unit test: contact profile

---

## Task 5.4: Online Status via WebSocket

**Input:** WebSocket Hub dari Phase 03, Task 5.2
**Output:** Real-time online/offline status broadcast

### Steps:
1. Update WebSocket Hub:
   - Saat client connects → broadcast `online_status` ke semua contacts
   - Saat client disconnects → broadcast `offline_status` ke semua contacts
   - `online_status` event:
     ```json
     {
       "type": "online_status",
       "payload": {
         "userId": "...",
         "isOnline": true,
         "lastSeen": "2026-02-11T10:30:00Z"
       }
     }
     ```
2. Get user's contacts → broadcast to each contact's user room
3. Debounce disconnect: wait 5 seconds before broadcasting offline
   - Prevents flicker on temporary disconnects
4. Store online status in Redis:
   - Key: `online:{userID}`, TTL: 5 menit
   - Refresh on every WS pong

### Acceptance Criteria:
- [x] Connect → contacts see user online
- [x] Disconnect → contacts see user offline (after 5s debounce)
- [x] LastSeen updated on disconnect
- [x] Redis tracks online status
- [x] No flicker on temporary disconnects

### Testing:
- [x] Unit test: connect broadcast
- [x] Unit test: disconnect broadcast (with debounce)
- [x] Unit test: reconnect cancels offline (debounce)
- [x] Unit test: status change visible to contacts

---

## Phase 05 Review

### Testing Checklist:
- [x] Profile setup: new user sets name + avatar
- [x] Profile update: change name, avatar, status
- [x] Contact sync: phone hashes → matched users
- [x] Contact list: sorted, with online status
- [x] Search: find user by phone number
- [x] Online status: real-time via WebSocket
- [x] Last seen: accurate timestamp
- [x] Delete account: all data removed
- [x] `go test ./...` — semua test pass

### Review Checklist:
- [x] Contact system sesuai `spesifikasi-chatat.md` section 2.3
- [x] Privacy: phone hash matching, no plain numbers exposed
- [x] Online status sesuai WA behavior
- [x] Error handling sesuai `docs/error-handling.md`
- [x] Naming sesuai `docs/naming-conventions.md`
- [x] Commit: `feat(user): implement user profile and contact system`
