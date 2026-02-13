# Phase 10: Topic System

> Implementasi topik sebagai ruang diskusi terfokus dari chat personal/grup.
> Phase ini menghasilkan fitur topik end-to-end.

**Estimasi:** 4 hari
**Dependency:** Phase 08 (Chat Group), Phase 09 (Real-time Messaging)
**Output:** Topic CRUD, messaging, dan UI lengkap.

---

## Task 10.1: Topic Service (Backend)

**Input:** Topic repository dari Phase 02
**Output:** Business logic untuk topic management

### Steps:
1. Buat `internal/service/topic_service.go`:
   ```go
   type TopicService interface {
       CreateTopic(ctx context.Context, input CreateTopicInput) (*model.Topic, error)
       GetTopic(ctx context.Context, topicID uuid.UUID) (*TopicDetail, error)
       ListByChat(ctx context.Context, chatID uuid.UUID) ([]*TopicListItem, error)
       ListByUser(ctx context.Context, userID uuid.UUID) ([]*TopicListItem, error)
       UpdateTopic(ctx context.Context, topicID uuid.UUID, input UpdateTopicInput) (*model.Topic, error)
       AddMember(ctx context.Context, topicID, userID, addedBy uuid.UUID) error
       RemoveMember(ctx context.Context, topicID, userID, removedBy uuid.UUID) error
       DeleteTopic(ctx context.Context, topicID, userID uuid.UUID) error
   }

   type CreateTopicInput struct {
       Name        string      `json:"name" validate:"required,min=1,max=100"`
       Icon        string      `json:"icon" validate:"required"`
       Description string      `json:"description"`
       ParentID    uuid.UUID   `json:"parentId" validate:"required"`
       MemberIDs   []uuid.UUID `json:"memberIds"`
   }

   type TopicListItem struct {
       Topic        model.Topic `json:"topic"`
       LastMessage  *model.TopicMessage `json:"lastMessage"`
       UnreadCount  int         `json:"unreadCount"`
       MemberCount  int         `json:"memberCount"`
   }

   type TopicDetail struct {
       Topic    model.Topic       `json:"topic"`
       Members  []*MemberInfo     `json:"members"`
       Parent   *model.Chat       `json:"parent"`
       DocCount int               `json:"docCount"`
   }
   ```
2. Implementasi CreateTopic:
   - Validate: parentID refers to a valid chat
   - Validate: creator must be member of parent chat
   - Validate: all memberIDs must be members of parent chat
   - If parent is personal chat → members = both participants (auto)
   - If parent is group → members selected from group members
   - Create topic + add members
   - Send system message to parent chat: "Andi membuat topik Pembagian Lahan"
   - Create topic room in WS hub
3. Implementasi member management:
   - Only add members who are also in parent chat
   - Admin: topic creator
   - System messages for add/remove

### Acceptance Criteria:
- [x] Topic created from personal chat: auto include both users
- [x] Topic created from group: select subset/all members
- [x] Members must be from parent chat
- [x] System messages in both parent chat and topic
- [x] Topic list sorted by last message
- [x] Unread count per topic
- [x] Delete topic by admin

### Testing:
- [x] Unit test: create from personal (auto members)
- [x] Unit test: create from group (subset members)
- [x] Unit test: invalid member (not in parent)
- [x] Unit test: add/remove member
- [x] Unit test: list topics by chat
- [x] Unit test: delete topic

---

## Task 10.2: Topic Message Service (Backend)

**Input:** Task 10.1
**Output:** Messaging within topics

### Steps:
1. Buat `internal/service/topic_message_service.go`:
   - Same interface pattern as MessageService
   - SendTopicMessage, GetTopicMessages, DeleteTopicMessage
2. WebSocket integration:
   - Topic room: `topic:{topicID}`
   - Auto-join users to their topic rooms on WS connect
   - Broadcast messages to topic room
3. Topic message status tracking (same as chat messages)
4. Read receipts for topic messages

### Acceptance Criteria:
- [x] Topic message CRUD same pattern as chat messages
- [x] Real-time delivery via WS topic rooms
- [x] Read receipts work in topics
- [x] Cursor pagination for topic messages

### Testing:
- [x] Unit test: send topic message
- [x] Unit test: topic message pagination
- [x] Integration test: real-time topic messaging

---

## Task 10.3: Topic Handler & Endpoints

**Input:** Task 10.1, 10.2
**Output:** REST endpoints untuk topics

### Steps:
1. Buat `internal/handler/topic_handler.go`:
   - `POST /api/v1/topics` → create topic
     - Body: `{"name": "Pembagian Lahan", "icon": "🌾", "parentId": "uuid", "memberIds": ["uuid"]}`
   - `GET /api/v1/topics/:topicId` → get topic detail
   - `GET /api/v1/chats/:chatId/topics` → list topics in chat
   - `GET /api/v1/topics` → list all user's topics
   - `PUT /api/v1/topics/:topicId` → update topic
   - `POST /api/v1/topics/:topicId/members` → add member
   - `DELETE /api/v1/topics/:topicId/members/:userId` → remove member
   - `DELETE /api/v1/topics/:topicId` → delete topic
2. Buat topic message endpoints:
   - `GET /api/v1/topics/:topicId/messages` → get messages
   - `POST /api/v1/topics/:topicId/messages` → send message
   - `DELETE /api/v1/topics/:topicId/messages/:messageId` → delete
3. Authorization:
   - Only topic members can access topic and messages
   - Only admin can update/delete topic
   - Only admin can add/remove members

### Acceptance Criteria:
- [x] All CRUD endpoints functioning
- [x] Authorization enforced
- [x] Message endpoints mirroring chat pattern
- [x] Consistent error responses

### Testing:
- [x] Integration test: create topic with members
- [x] Integration test: topic messaging
- [x] Integration test: member management

---

## Task 10.4: Topic UI (Frontend)

**Input:** Navigation, chat screens, WebSocket
**Output:** Topic creation, list, and discussion screens

### Steps:
1. Buat `src/screens/topic/CreateTopicScreen.tsx`:
   - When from personal chat:
     - Both users auto-selected (shown but not removable)
     - Topic name input
     - Topic icon picker (emoji grid from spec 4.3.1)
     - Description (optional)
     - "Buat Topik" button
   - When from group:
     - Member selection (checkboxes, from group members)
     - "Pilih Semua" option
     - Topic name + icon + description
     - "Buat Topik" button
2. Buat `src/screens/topic/TopicListScreen.tsx`:
   - Accessed from Tab Topik (📌) in group chat
   - FlatList of topics:
     - Topic icon + name
     - Last message preview
     - Unread badge
     - Member count
   - "Buat Topik" FAB
   - Empty state: "Belum ada topik. Buat topik untuk diskusi terfokus."
3. Buat `src/screens/topic/TopicScreen.tsx`:
   - Header: topic icon + name, member count
   - 2 tabs:
     - **Tab Diskusi (💬)**: messages (same as ChatScreen)
     - **Tab Dokumen (📄)**: document list in topic (Phase 13+)
   - Same message features as chat (reply, delete, etc.)
   - Topic info accessible from header tap
4. Buat `src/screens/topic/TopicInfoScreen.tsx`:
   - Topic icon, name, description (editable by admin)
   - Parent chat info
   - Member list
   - Admin actions (add/remove member from parent members)
   - "Hapus Topik" button (admin only)
5. Buat `src/stores/topicStore.ts`:
   ```tsx
   interface TopicState {
     topics: Record<string, TopicListItem[]>; // chatId → topics
     fetchTopics: (chatId: string) => Promise<void>;
     addTopic: (chatId: string, topic: TopicListItem) => void;
   }
   ```
6. WebSocket integration:
   - Listen for topic messages
   - Update unread counts
   - Real-time typing in topics

### Acceptance Criteria:
- [x] Create from personal: auto 2 members
- [x] Create from group: select from group members
- [x] Topic icon selection from predefined emoji list
- [x] Topic list with unread badges
- [x] Topic discussion: same UX as chat
- [x] Topic info: member management
- [x] Real-time messaging in topics

### Testing:
- [x] Component test: CreateTopicScreen (personal vs group)
- [x] Component test: TopicListScreen
- [x] Component test: TopicScreen (tab switching)
- [x] Store test: topicStore
- [x] Integration test: create topic → send message

---

## Phase 10 Review

### Testing Checklist:
- [x] Backend: create topic from personal/group
- [x] Backend: topic membership validation
- [x] Backend: topic messaging
- [x] Frontend: create topic wizard
- [x] Frontend: topic list with badges
- [x] Frontend: topic discussion screen
- [x] Frontend: topic info + member management
- [x] Real-time: messages delivered in topics
- [x] System messages in parent chat
- [x] `go test ./...` + `npm test` pass

### Review Checklist:
- [x] Topic sesuai `spesifikasi-chatat.md` section 4
- [x] Membership rules enforced (from parent only)
- [x] Topic icon list sesuai spec 4.3.1
- [x] Tab layout: Diskusi + Dokumen
- [x] Commit: `feat(topic): implement topic system with discussion`
