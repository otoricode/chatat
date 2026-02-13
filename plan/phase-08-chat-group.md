# Phase 08: Chat Group

> Implementasi chat grup — CRUD grup, manajemen anggota, dan UI grup.
> Phase ini menghasilkan fitur grup identik dengan WhatsApp group.

**Estimasi:** 4 hari
**Dependency:** Phase 07 (Chat Personal)
**Output:** Group chat berfungsi: create, manage members, chat grup, group info.

---

## Task 8.1: Group Service (Backend)

**Input:** Chat service dari Phase 07
**Output:** Business logic untuk group chat

### Steps:
1. Extend `internal/service/chat_service.go` atau buat `group_service.go`:
   ```go
   type GroupService interface {
       CreateGroup(ctx context.Context, input CreateGroupInput) (*model.Chat, error)
       UpdateGroup(ctx context.Context, chatID uuid.UUID, input UpdateGroupInput) (*model.Chat, error)
       AddMember(ctx context.Context, chatID, userID, addedBy uuid.UUID) error
       RemoveMember(ctx context.Context, chatID, userID, removedBy uuid.UUID) error
       PromoteToAdmin(ctx context.Context, chatID, userID, promotedBy uuid.UUID) error
       LeaveGroup(ctx context.Context, chatID, userID uuid.UUID) error
       DeleteGroup(ctx context.Context, chatID, userID uuid.UUID) error
       GetGroupInfo(ctx context.Context, chatID uuid.UUID) (*GroupInfo, error)
   }

   type CreateGroupInput struct {
       Name        string      `json:"name" validate:"required,min=1,max=100"`
       Icon        string      `json:"icon" validate:"required"`
       Description string      `json:"description"`
       MemberIDs   []uuid.UUID `json:"memberIds" validate:"required,min=2"`
   }

   type UpdateGroupInput struct {
       Name        *string `json:"name"`
       Icon        *string `json:"icon"`
       Description *string `json:"description"`
   }

   type GroupInfo struct {
       Chat        model.Chat    `json:"chat"`
       Members     []*MemberInfo `json:"members"`
       TopicCount  int           `json:"topicCount"`
       DocCount    int           `json:"docCount"`
   }

   type MemberInfo struct {
       User     model.User `json:"user"`
       Role     string     `json:"role"`
       IsOnline bool       `json:"isOnline"`
       JoinedAt time.Time  `json:"joinedAt"`
   }
   ```
2. Implementasi CreateGroup:
   - Validate: min 2 members (+ creator = min 3 total)
   - Create chat with type="group"
   - Add creator as admin
   - Add all memberIDs as members
   - Send system message: "Andi membuat grup"
   - Broadcast via WebSocket: group created
3. Implementasi member management:
   - AddMember: only admin can add, send system msg "Andi menambahkan Budi"
   - RemoveMember: only admin can remove (not self), system msg
   - PromoteToAdmin: only admin can promote
   - LeaveGroup: any member, system msg "Andi keluar dari grup"
   - Creator cannot be removed or leave
4. Implementasi DeleteGroup:
   - Only creator (first admin) can delete
   - CASCADE: delete messages, members, topics, documents via FK

### Acceptance Criteria:
- [x] Create group: min 3 people (creator + 2)
- [x] Group has name, icon (emoji), optional description
- [x] Admin can: add/remove members, update group info
- [x] System messages for membership changes
- [x] Creator cannot be removed
- [x] Leave group notification
- [x] Group delete cascades all related data

### Testing:
- [x] Unit test: create group (valid + invalid member count)
- [x] Unit test: add member (admin vs non-admin)
- [x] Unit test: remove member
- [x] Unit test: promote to admin
- [x] Unit test: leave group
- [x] Unit test: delete group (admin vs non-admin)
- [x] Unit test: system messages generated

---

## Task 8.2: Group Handler & Endpoints

**Input:** Task 8.1
**Output:** REST endpoints untuk group management

### Steps:
1. Extend `internal/handler/chat_handler.go`:
   - `POST /api/v1/chats/group` → create group
     - Body: `{"name": "Tim Proyek", "icon": "💼", "memberIds": ["uuid1", "uuid2"]}`
   - `PUT /api/v1/chats/:chatId` → update group info
     - Body: `{"name": "Tim Proyek v2", "icon": "🚀"}`
   - `GET /api/v1/chats/:chatId/info` → get group info + members
   - `POST /api/v1/chats/:chatId/members` → add member
     - Body: `{"userId": "uuid"}`
   - `DELETE /api/v1/chats/:chatId/members/:userId` → remove member
   - `PUT /api/v1/chats/:chatId/members/:userId/admin` → promote to admin
   - `POST /api/v1/chats/:chatId/leave` → leave group
   - `DELETE /api/v1/chats/:chatId` → delete group
2. Authorization:
   - Update/delete group: admin only
   - Add/remove members: admin only
   - Leave: any member
   - View info: any member
3. WebSocket broadcast for group events:
   - `group_updated`: name/icon changed
   - `member_added`: new member joins
   - `member_removed`: member kicked/left

### Acceptance Criteria:
- [x] All CRUD endpoints functioning
- [x] Admin-only actions enforced
- [x] WebSocket broadcasts for group events
- [x] Error responses for unauthorized actions

### Testing:
- [x] Integration test: create group flow
- [x] Integration test: member management
- [x] Integration test: authorization enforcement

---

## Task 8.3: Create Group Screen (Frontend)

**Input:** Contact list, navigation
**Output:** Group creation wizard

### Steps:
1. Buat `src/screens/chat/CreateGroupScreen.tsx`:
   - Step 1: Select members
     - Search bar for contacts
     - Contact list with checkboxes
     - Selected members shown as chips at top
     - "Selanjutnya" button (enabled when >= 2 selected)
   - Step 2: Group details
     - Emoji picker for group icon
     - Group name input (required)
     - Description input (optional)
     - "Buat Grup" button
2. Buat `src/components/chat/MemberChip.tsx`:
   - Show avatar + name + X button
3. After creation: navigate to new group chat screen

### Acceptance Criteria:
- [x] Multi-select contacts (min 2)
- [x] Selected members shown as chips
- [x] Search filtering contacts
- [x] Group name + icon required
- [x] Create → navigate to group chat
- [x] Loading state during creation

### Testing:
- [x] Component test: contact selection
- [x] Component test: member chips
- [x] Component test: group creation form validation

---

## Task 8.4: Group Chat Screen (Frontend)

**Input:** Chat screen dari Phase 07
**Output:** Group-specific chat screen with 3 tabs

### Steps:
1. Extend ChatScreen for group-specific features:
   - Header: group icon + group name + member count
   - Tap header → GroupInfoScreen
   - Message bubbles show sender name (color-coded) + avatar
   - 3 tabs within group chat:
     - **Tab Chat (💬)**: messages + inline document cards
     - **Tab Dokumen (📄)**: document list (Phase 13+)
     - **Tab Topik (📌)**: topic list (Phase 10)
2. Buat tab container:
   ```tsx
   // TopTabNavigator within GroupChatScreen
   <Tab.Navigator>
     <Tab.Screen name="Chat" component={GroupChatTab} />
     <Tab.Screen name="Dokumen" component={GroupDocumentsTab} />
     <Tab.Screen name="Topik" component={GroupTopicsTab} />
   </Tab.Navigator>
   ```
3. Sender name colors:
   - Assign consistent color per user from a palette
   - Color palette: purple, blue, teal, orange, pink, cyan
4. "Buat Topik" button di Tab Topik → CreateTopicScreen (Phase 10)
5. Document/Topic tab: placeholder "Segera hadir" until Phase 10/13

### Acceptance Criteria:
- [x] Group header: icon, name, member count
- [x] Sender names visible (colored) above bubbles
- [x] 3 tabs switching
- [x] Tab badges: unread count on Chat tab
- [x] Tap header → group info

### Testing:
- [x] Component test: group header
- [x] Component test: sender name rendering
- [x] Component test: tab switching

---

## Task 8.5: Group Info Screen (Frontend)

**Input:** Task 8.2 (group endpoints)
**Output:** Group settings and member management screen

### Steps:
1. Buat `src/screens/chat/GroupInfoScreen.tsx`:
   - Group icon (large, editable by admin)
   - Group name (editable by admin)
   - Description (editable by admin)
   - Member count
   - Member list:
     - Avatar + Name + Role badge (Admin)
     - Online indicator
     - Tap member (if admin): promote/remove options
   - "Tambah Anggota" button (admin only)
   - "Keluar dari Grup" button (red)
   - "Hapus Grup" button (admin only, red)
2. Admin actions via BottomSheet:
   - Promote to admin
   - Remove from group
3. Leave group confirmation dialog
4. Delete group confirmation dialog

### Acceptance Criteria:
- [x] Group details editable by admin
- [x] Member list with roles
- [x] Admin actions: add, remove, promote
- [x] Leave group with confirmation
- [x] Delete group with confirmation (admin only)
- [x] Non-admin: no edit buttons

### Testing:
- [x] Component test: member list rendering
- [x] Component test: admin vs non-admin view
- [x] Component test: leave confirmation
- [x] Integration test: update group info

---

## Phase 08 Review

### Testing Checklist:
- [x] Backend: create group with members
- [x] Backend: add/remove members (admin)
- [x] Backend: leave group
- [x] Backend: system messages generated
- [x] Frontend: create group wizard
- [x] Frontend: group chat with sender names
- [x] Frontend: 3-tab layout
- [x] Frontend: group info + member management
- [x] End-to-end: create group → chat → manage members
- [x] `go test ./...` + `npm test` pass

### Review Checklist:
- [x] Group sesuai `spesifikasi-chatat.md` section 3.3
- [x] Tab layout sesuai spec
- [x] Admin rules enforced
- [x] System messages clear and accurate
- [x] Commit: `feat(chat): implement group chat with member management`
