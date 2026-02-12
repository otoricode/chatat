# Phase 09: Real-time Messaging

> Implementasi WebSocket integration untuk messaging real-time.
> Phase ini menghasilkan typing indicators, read receipts, dan online status.

**Estimasi:** 4 hari
**Dependency:** Phase 07 (Chat Personal), Phase 08 (Chat Group)
**Output:** Real-time messaging: delivery status, typing, online, read receipts.

---

## Task 9.1: WebSocket Client (React Native)

**Input:** WebSocket hub dari Phase 03
**Output:** React Native WebSocket client dengan auto-reconnect

### Steps:
1. Buat `src/services/ws/WebSocketClient.ts`:
   ```tsx
   class WebSocketClient {
     private ws: WebSocket | null = null;
     private reconnectAttempts = 0;
     private maxReconnectAttempts = 10;
     private reconnectDelay = 1000; // doubles each attempt
     private messageHandlers: Map<string, Set<(payload: any) => void>>;
     private pendingMessages: WSMessage[]; // offline queue

     connect(url: string, token: string): void
     disconnect(): void
     send(type: string, payload: any): void
     on(type: string, handler: (payload: any) => void): () => void
     off(type: string, handler: (payload: any) => void): void

     // Internal
     private handleOpen(): void
     private handleMessage(event: MessageEvent): void
     private handleClose(event: CloseEvent): void
     private handleError(event: Event): void
     private reconnect(): void
     private flushPendingMessages(): void
   }
   ```
2. Implementasi:
   - Connect: `ws://host/ws?token=<accessToken>`
   - Auto-reconnect: exponential backoff (1s, 2s, 4s, 8s... max 30s)
   - Offline queue: messages sent while disconnected → send on reconnect
   - Ping/pong: keep connection alive
   - Event emitter pattern: handlers per message type
3. Buat `src/services/ws/index.ts`:
   ```tsx
   export const wsClient = new WebSocketClient();
   ```
4. Buat `src/hooks/useWebSocket.ts`:
   ```tsx
   export function useWebSocket() {
     const token = useAuthStore(state => state.accessToken);

     useEffect(() => {
       if (token) {
         wsClient.connect(Config.WS_URL, token);
       }
       return () => wsClient.disconnect();
     }, [token]);
   }

   export function useWSEvent<T>(type: string, handler: (payload: T) => void) {
     useEffect(() => {
       return wsClient.on(type, handler);
     }, [type, handler]);
   }
   ```

### Acceptance Criteria:
- [ ] WebSocket connects with auth token
- [ ] Auto-reconnect with exponential backoff
- [ ] Offline message queue flushed on reconnect
- [ ] Event system: subscribe/unsubscribe per type
- [ ] Ping/pong keepalive
- [ ] Clean disconnect on logout
- [ ] Connection state observable

### Testing:
- [ ] Unit test: connect/disconnect
- [ ] Unit test: message sending
- [ ] Unit test: event handlers
- [ ] Unit test: reconnect logic
- [ ] Unit test: offline queue

---

## Task 9.2: Real-time Message Delivery

**Input:** Task 9.1, Message service dari Phase 07
**Output:** Messages delivered in real-time via WebSocket

### Steps:
1. Update backend message handling:
   - Saat message created → broadcast ke chat room via Hub
   - Message event:
     ```json
     {
       "type": "message",
       "payload": {
         "chatId": "uuid",
         "message": { "id": "uuid", "senderId": "uuid", "content": "Hello", ... }
       }
     }
     ```
2. Update frontend:
   - Listen for `message` events
   - Add to messageStore
   - Update chatStore last message + unread count
   - If chat is open → auto-scroll to new message
   - If chat is not open → increment unread badge
   - Play notification sound (if app is in foreground, different chat)
3. Buat `src/hooks/useMessageListener.ts`:
   ```tsx
   export function useMessageListener() {
     const addMessage = useMessageStore(s => s.addMessage);
     const updateLastMessage = useChatStore(s => s.updateLastMessage);

     useWSEvent('message', (payload: MessageEvent) => {
       addMessage(payload.chatId, payload.message);
       updateLastMessage(payload.chatId, payload.message);
     });
   }
   ```
4. Same pattern for topic messages (separate event type or flag)

### Acceptance Criteria:
- [ ] Message sent → appears instantly on receiver's screen
- [ ] Chat list: last message updates real-time
- [ ] Unread count increments for non-active chats
- [ ] Auto-scroll to new message in active chat
- [ ] Works for personal + group + topic messages

### Testing:
- [ ] Integration test: send message → receive via WS
- [ ] Unit test: message listener updates stores
- [ ] Unit test: unread count increment

---

## Task 9.3: Delivery & Read Receipts

**Input:** Message status dari Phase 02, WebSocket
**Output:** ✓ sent, ✓✓ delivered, blue ✓✓ read

### Steps:
1. Backend status updates:
   - **Sent (✓)**: message stored in DB → status = 'sent'
   - **Delivered (✓✓)**: receiver's device acknowledges receipt → status = 'delivered'
   - **Read (blue ✓✓)**: receiver opens chat → status = 'read'
2. Delivery acknowledgement:
   - Client receives message → sends ack back via WS:
     ```json
     {"type": "message_ack", "payload": {"messageId": "uuid", "status": "delivered"}}
     ```
   - Server updates message_status → broadcast to sender
3. Read receipts:
   - When user opens a chat → mark all messages as read
   - Send batch read receipt via WS:
     ```json
     {"type": "read_receipt", "payload": {"chatId": "uuid", "lastReadMessageId": "uuid"}}
     ```
   - Server: update all messages in chat up to lastReadMessageId as 'read'
   - Broadcast read receipt to other members
4. Status change event:
   ```json
   {
     "type": "message_status",
     "payload": {
       "chatId": "uuid",
       "messageId": "uuid",
       "userId": "uuid",
       "status": "read"
     }
   }
   ```
5. Frontend rendering:
   - ✓ (single check, gray): sent
   - ✓✓ (double check, gray): delivered
   - ✓✓ (double check, blue): read by all
   - For groups: read when ALL members have read
   - Show in MessageBubble component

### Acceptance Criteria:
- [ ] Sent status: immediately on send
- [ ] Delivered status: when receiver's device gets message
- [ ] Read status: when receiver opens chat
- [ ] Group: read = all members have read
- [ ] Status icons render correctly (✓, ✓✓, blue ✓✓)
- [ ] Batch read receipt (not per-message)
- [ ] Status transitions are one-way (sent→delivered→read)

### Testing:
- [ ] Unit test: status transitions
- [ ] Unit test: delivery ack flow
- [ ] Unit test: read receipt flow
- [ ] Unit test: group read status (all members)
- [ ] Component test: status icon rendering
- [ ] Integration test: full delivery cycle

---

## Task 9.4: Typing Indicators

**Input:** WebSocket hub, Redis
**Output:** "sedang mengetik..." indicator

### Steps:
1. Backend:
   - Receive typing event via WS:
     ```json
     {"type": "typing", "payload": {"chatId": "uuid", "isTyping": true}}
     ```
   - Store in Redis: key `typing:{chatId}:{userId}`, TTL 3 detik
   - Broadcast to chat room (exclude sender)
   - Typing event:
     ```json
     {
       "type": "typing",
       "payload": {"chatId": "uuid", "userId": "uuid", "userName": "Andi", "isTyping": true}
     }
     ```
2. Frontend sending:
   - Debounce: send typing=true saat user mulai mengetik
   - Send typing=false saat user berhenti (2 detik idle) atau sends message
   - Max 1 typing event per 2 detik
3. Frontend receiving:
   - Show "Andi sedang mengetik..." di chat header
   - For groups: "Andi, Budi sedang mengetik..." (max 2 names)
   - Auto-clear after 3 detik (if no update received)
4. Update chat list:
   - Show "sedang mengetik..." instead of last message preview
   - Green italic text

### Acceptance Criteria:
- [ ] Typing indicator shows when contact types
- [ ] Auto-clear after 3 seconds
- [ ] Debounced sending (max 1 per 2s)
- [ ] Group: show up to 2 names
- [ ] Chat list: typing preview replaces last message
- [ ] No phantom typing (always clears)

### Testing:
- [ ] Unit test: typing debounce
- [ ] Unit test: typing auto-clear
- [ ] Unit test: group typing display
- [ ] Component test: typing indicator UI
- [ ] Integration test: typing flow between users

---

## Task 9.5: Online Status Real-time

**Input:** WebSocket, Online status dari Phase 05
**Output:** Real-time online/offline updates

### Steps:
1. Integrate with contact store:
   - Listen for `online_status` WS events
   - Update contact online status in store
   - Update chat list items (online indicator)
2. Chat header status:
   - Online: "online" (green text)
   - Offline: "terakhir dilihat pukul HH:MM" atau "terakhir dilihat kemarin pukul HH:MM"
   - Format: relative ("baru saja", "5 menit yang lalu") jika < 1 jam
3. Contact list: sort online contacts first

### Acceptance Criteria:
- [ ] Online status updates in real-time
- [ ] Chat header shows accurate status
- [ ] Contact list sorts online first
- [ ] Last seen format: relative + absolute
- [ ] Status updates debounced (no flicker)

### Testing:
- [ ] Unit test: status update handler
- [ ] Unit test: last seen formatting
- [ ] Component test: status text rendering
- [ ] Integration test: connect → contacts see online

---

## Phase 09 Review

### Testing Checklist:
- [ ] WebSocket: connect, reconnect, offline queue
- [ ] Messages: real-time delivery both directions
- [ ] Delivery status: ✓ → ✓✓ → blue ✓✓
- [ ] Typing: show/hide correctly, debounced
- [ ] Online: status changes reflect immediately
- [ ] Group: all features work with multiple users
- [ ] Reconnect: messages delivered after reconnect
- [ ] `go test ./...` + `npm test` pass

### Review Checklist:
- [ ] Real-time features sesuai `spesifikasi-chatat.md` section 3.4
- [ ] Status icons sesuai WA behavior
- [ ] No memory leaks on WS reconnects
- [ ] No goroutine leaks on server
- [ ] Event types documented
- [ ] Commit: `feat(ws): implement real-time messaging and status`
