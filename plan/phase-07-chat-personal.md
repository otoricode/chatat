# Phase 07: Chat Personal

> Implementasi chat personal (1-on-1) — backend API dan frontend screens.
> Phase ini menghasilkan fitur chat personal yang identik dengan WhatsApp.

**Estimasi:** 5 hari
**Dependency:** Phase 03 (API), Phase 05 (User & Contact), Phase 06 (Mobile Shell)
**Output:** Chat personal berfungsi end-to-end: daftar chat, kirim pesan, reply, delete.

---

## Task 7.1: Chat Service (Backend)

**Input:** Chat repository dari Phase 02
**Output:** Business logic untuk personal chat

### Steps:
1. Buat `internal/service/chat_service.go`:
   ```go
   type ChatService interface {
       CreatePersonalChat(ctx context.Context, userID, contactID uuid.UUID) (*model.Chat, error)
       GetOrCreatePersonalChat(ctx context.Context, userID, contactID uuid.UUID) (*model.Chat, error)
       ListChats(ctx context.Context, userID uuid.UUID) ([]*ChatListItem, error)
       GetChat(ctx context.Context, chatID, userID uuid.UUID) (*ChatDetail, error)
       PinChat(ctx context.Context, chatID, userID uuid.UUID) error
       UnpinChat(ctx context.Context, chatID, userID uuid.UUID) error
       ArchiveChat(ctx context.Context, chatID, userID uuid.UUID) error
   }

   type ChatListItem struct {
       Chat         model.Chat    `json:"chat"`
       LastMessage   *model.Message `json:"lastMessage"`
       UnreadCount   int           `json:"unreadCount"`
       OtherUser     *model.User   `json:"otherUser"` // for personal chats
       IsOnline      bool          `json:"isOnline"`
   }

   type ChatDetail struct {
       Chat    model.Chat       `json:"chat"`
       Members []*model.User    `json:"members"`
   }
   ```
2. Implementasi GetOrCreatePersonalChat:
   - Check jika personal chat antara kedua user sudah exist
   - Jika ada → return existing
   - Jika tidak → create baru + add both as members
3. Implementasi ListChats:
   - Query: semua chat dimana user adalah member
   - Join: last message per chat
   - Join: unread count per chat
   - Sort: pinned first, then by last_message.created_at DESC
   - Include: other user info (for personal chats)
4. Implementasi pin/archive (update sort preference)

### Acceptance Criteria:
- [ ] Personal chat: get-or-create pattern (idempotent)
- [ ] Chat list: sorted by last message, pinned first
- [ ] Unread count accurate per chat
- [ ] Last message preview included
- [ ] Other user info included for personal chats

### Testing:
- [ ] Unit test: create personal chat
- [ ] Unit test: get existing personal chat
- [ ] Unit test: list chats (sort order)
- [ ] Unit test: unread count calculation
- [ ] Unit test: pin/unpin

---

## Task 7.2: Message Service (Backend)

**Input:** Message repository dari Phase 02
**Output:** Business logic untuk messaging

### Steps:
1. Buat `internal/service/message_service.go`:
   ```go
   type MessageService interface {
       SendMessage(ctx context.Context, input SendMessageInput) (*model.Message, error)
       GetMessages(ctx context.Context, chatID uuid.UUID, cursor string, limit int) (*MessagePage, error)
       ForwardMessage(ctx context.Context, messageID, senderID, targetChatID uuid.UUID) (*model.Message, error)
       DeleteMessage(ctx context.Context, messageID, userID uuid.UUID, forAll bool) error
       SearchMessages(ctx context.Context, chatID uuid.UUID, query string) ([]*model.Message, error)
   }

   type SendMessageInput struct {
       ChatID    uuid.UUID `json:"chatId"`
       SenderID  uuid.UUID `json:"-"` // from auth context
       Content   string    `json:"content"`
       ReplyToID *uuid.UUID `json:"replyToId"`
       Type      string    `json:"type"` // text, image, file, document_card
       Metadata  *json.RawMessage `json:"metadata"`
   }

   type MessagePage struct {
       Messages []*model.Message `json:"messages"`
       Cursor   string           `json:"cursor"`
       HasMore  bool             `json:"hasMore"`
   }
   ```
2. Implementasi SendMessage:
   - Validate: user is member of chat
   - Validate: content not empty (for text)
   - Validate: replyToID exists in same chat (if provided)
   - Create message in DB
   - Create message_status entries for all other members (status: 'sent')
   - Broadcast via WebSocket to chat room
   - Return created message
3. Implementasi GetMessages:
   - Cursor-based pagination (created_at based)
   - Default limit: 50
   - Include: sender info (name, avatar)
   - Include: reply-to message preview (if reply)
   - Exclude: messages deleted for requesting user
4. Implementasi ForwardMessage:
   - Validate: user is member of target chat
   - Create new message in target chat with original content
   - Set metadata: `{"forwarded": true, "originalChatId": "...", "originalMessageId": "..."}`
   - Broadcast to target chat via WebSocket
5. Implementasi DeleteMessage:
   - `forAll = true`: mark `deleted_for_all = true` → all users see "Pesan dihapus"
   - `forAll = false`: only hide for requesting user
   - Only sender can delete for all (within 1 hour)

### Acceptance Criteria:
- [ ] Send message: validated, stored, broadcast
- [ ] Cursor pagination: 50 per page, ordered by created_at DESC
- [ ] Reply: includes original message preview
- [ ] Forward: copies message to target chat with forwarded metadata
- [ ] Delete for self: hidden only for requester
- [ ] Delete for all: all users see "Pesan dihapus"
- [ ] Delete for all: only within 1 hour by sender

### Testing:
- [ ] Unit test: send message (valid)
- [ ] Unit test: send message (not a member → error)
- [ ] Unit test: cursor pagination
- [ ] Unit test: reply message
- [ ] Unit test: forward message to another chat
- [ ] Unit test: forward message (not a member of target → error)
- [ ] Unit test: delete for self
- [ ] Unit test: delete for all
- [ ] Unit test: delete for all expired (>1 hour)

---

## Task 7.3: Chat Handler & Endpoints

**Input:** Task 7.1, 7.2
**Output:** REST endpoints untuk chat dan messages

### Steps:
1. Buat `internal/handler/chat_handler.go`:
   - `GET /api/v1/chats` → list user's chats
   - `POST /api/v1/chats/personal` → create/get personal chat
     - Body: `{"contactId": "uuid"}`
   - `GET /api/v1/chats/:chatId` → get chat detail
   - `PUT /api/v1/chats/:chatId/pin` → pin chat
   - `DELETE /api/v1/chats/:chatId/pin` → unpin chat
   - `PUT /api/v1/chats/:chatId/archive` → archive chat
2. Buat `internal/handler/message_handler.go`:
   - `GET /api/v1/chats/:chatId/messages?cursor=xxx&limit=50` → get messages
   - `POST /api/v1/chats/:chatId/messages` → send message
     - Body: `{"content": "Hello", "replyToId": null, "type": "text"}`
   - `DELETE /api/v1/chats/:chatId/messages/:messageId` → delete message
     - Query: `?forAll=true`
   - `POST /api/v1/chats/:chatId/messages/:messageId/forward` → forward message
     - Body: `{"targetChatId": "uuid"}`
   - `GET /api/v1/chats/:chatId/messages/search?q=keyword` → search messages
3. Authorization checks:
   - All endpoints: verify user is member of chat
   - Delete for all: verify user is sender

### Acceptance Criteria:
- [ ] All endpoints return correct response format
- [ ] Authorization: only chat members can access
- [ ] Pagination: cursor + hasMore
- [ ] Error responses for invalid requests

### Testing:
- [ ] Integration test: create personal chat and send messages
- [ ] Integration test: pagination flow
- [ ] Integration test: unauthorized access blocked

---

## Task 7.4: Chat List Screen (Frontend)

**Input:** Task 6.1 navigation, API endpoints
**Output:** WhatsApp-style chat list screen

### Steps:
1. Buat `src/screens/chat/ChatListScreen.tsx`:
   - FlatList of chats
   - Each item:
     - Avatar (emoji of contact/group)
     - Name (contact name or group name)
     - Last message preview (truncated, sender name in group)
     - Timestamp (relative: "just now", "5m", "2j", atau tanggal)
     - Unread badge (green circle with count)
     - Pin icon (if pinned)
   - Pull-to-refresh
   - Empty state: "Belum ada chat. Tap + untuk memulai."
2. Buat `src/stores/chatStore.ts`:
   ```tsx
   interface ChatState {
     chats: ChatListItem[];
     isLoading: boolean;
     fetchChats: () => Promise<void>;
     updateLastMessage: (chatId: string, message: Message) => void;
     updateUnreadCount: (chatId: string, count: number) => void;
   }
   ```
3. Buat `src/hooks/useChats.ts`:
   - `useChats()` → fetch dan return chats
   - Auto-refresh via WebSocket events
4. Long-press actions (bottom sheet):
   - Pin / Unpin
   - Archive
   - Tandai Dibaca (mark all as read → API batch mark read)
5. FAB action: open ContactListScreen

### Acceptance Criteria:
- [ ] Chat list sorted: pinned first, then by last message time
- [ ] Unread badge shows count (99+ for > 99)
- [ ] Last message preview truncated
- [ ] Timestamp relative format
- [ ] Pull-to-refresh
- [ ] Long-press context menu
- [ ] FAB opens contact list
- [ ] Empty state

### Testing:
- [ ] Component test: ChatListItem renders correctly
- [ ] Component test: unread badge
- [ ] Component test: timestamp formatting
- [ ] Store test: fetchChats
- [ ] Store test: updateLastMessage

---

## Task 7.5: Chat Screen (Frontend)

**Input:** Task 7.4, WebSocket dari Phase 03
**Output:** Full chat screen identik WhatsApp

### Steps:
1. Buat `src/screens/chat/ChatScreen.tsx`:
   - Header:
     - Back arrow
     - Avatar + Name of contact
     - Status: "online" (green) atau "terakhir dilihat pukul HH:MM"
   - Message list (inverted FlatList):
     - Date separators ("Hari Ini", "Kemarin", "12 Feb 2026")
     - Message bubbles:
       - Own messages: right-aligned, green-ish background
       - Others: left-aligned, surface2 background
     - Timestamp HH:MM di setiap bubble
     - Status icons: ✓ sent, ✓✓ delivered, blue ✓✓ read
     - Reply preview (attached to bubble)
     - "Pesan dihapus" for deleted messages
   - Input bar:
     - Text input (multiline)
     - Emoji button (opens keyboard)
     - Attachment button (📎): buat dokumen, kirim foto
     - Send button (➤): green, appears when text not empty
   - Swipe right on message → reply
   - Long-press message → context menu:
     - Reply
     - Forward
     - Copy
     - Delete (for me / for everyone)
2. Buat `src/components/chat/MessageBubble.tsx`:
   - Props: message, isSelf, showSender (for groups)
   - Render: content + timestamp + status
   - Reply preview (if reply)
   - Document card preview (if type=document_card)
3. Buat `src/components/chat/ChatInput.tsx`:
   - Multiline TextInput
   - Send button visibility toggle
   - Attachment picker
4. Buat `src/stores/messageStore.ts`:
   ```tsx
   interface MessageState {
     messages: Record<string, Message[]>;  // chatId → messages
     isLoading: boolean;
     hasMore: Record<string, boolean>;
     cursor: Record<string, string>;
     fetchMessages: (chatId: string) => Promise<void>;
     fetchMore: (chatId: string) => Promise<void>;
     addMessage: (chatId: string, message: Message) => void;
     deleteMessage: (chatId: string, messageId: string) => void;
   }
   ```
5. Infinite scroll: load more saat scroll ke atas (older messages)

### Acceptance Criteria:
- [ ] Message bubbles: left/right aligned, correct colors
- [ ] Timestamps: HH:MM format on each bubble
- [ ] Date separators: auto-inserted between different days
- [ ] Status: ✓ / ✓✓ / blue ✓✓ indicators
- [ ] Reply: swipe-to-reply, preview shown
- [ ] Delete: soft delete ("Pesan dihapus")
- [ ] Input: multiline, send button, attachment button
- [ ] Infinite scroll (cursor pagination)
- [ ] Auto-scroll to bottom on new message
- [ ] Keyboard avoidance (input stays above keyboard)

### Testing:
- [ ] Component test: MessageBubble (self/other)
- [ ] Component test: ChatInput (send toggle)
- [ ] Component test: DateSeparator
- [ ] Store test: fetchMessages, addMessage
- [ ] Integration test: send and receive message flow

---

## Task 7.6: Contact List Screen (Frontend)

**Input:** Contact API dari Phase 05, Navigation
**Output:** Contact selection screen

### Steps:
1. Buat `src/screens/contact/ContactListScreen.tsx`:
   - Header: "Pilih Kontak"
   - Search bar (filter by name/number)
   - "Buat Grup Baru" button at top
   - FlatList of contacts:
     - Avatar (emoji)
     - Name (from phone contacts)
     - Status text (if set)
     - Online indicator (green dot)
   - Section headers: alphabetical (A, B, C...)
   - Tap contact → open/create personal chat
2. Buat `src/hooks/useContacts.ts`:
   - Request device contact permission
   - Read device contacts
   - Hash phone numbers
   - Sync with server
   - Return matched contacts with online status
3. Buat `src/stores/contactStore.ts`:
   ```tsx
   interface ContactState {
     contacts: ContactInfo[];
     isLoading: boolean;
     lastSynced: string | null;
     syncContacts: () => Promise<void>;
     searchContacts: (query: string) => ContactInfo[];
   }
   ```

### Acceptance Criteria:
- [ ] Contact permission requested
- [ ] Contacts synced with server
- [ ] Alphabetical sections
- [ ] Search filtering (client-side)
- [ ] Online indicator
- [ ] Tap → open personal chat
- [ ] "Buat Grup" button visible

### Testing:
- [ ] Component test: contact list rendering
- [ ] Component test: search filtering
- [ ] Store test: sync contacts
- [ ] Hook test: useContacts

---

## Phase 07 Review

### Testing Checklist:
- [ ] Backend: create personal chat API
- [ ] Backend: send/receive messages API
- [ ] Backend: cursor pagination
- [ ] Backend: reply, delete message
- [ ] Frontend: chat list displays correctly
- [ ] Frontend: chat screen WhatsApp-identical
- [ ] Frontend: contact list → start chat
- [ ] Frontend: real-time message (via WebSocket)
- [ ] End-to-end: User A sends → User B receives
- [ ] `go test ./...` + `npm test` pass

### Review Checklist:
- [ ] Chat sesuai `spesifikasi-chatat.md` section 3
- [ ] Bubble colors sesuai spec 9.2
- [ ] Message features sesuai spec 3.4
- [ ] Error handling sesuai `docs/error-handling.md`
- [ ] Commit: `feat(chat): implement personal chat with messaging`
