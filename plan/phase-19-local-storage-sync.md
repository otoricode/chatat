# Phase 19: Local Storage & Sync

> Implementasi local storage di mobile untuk offline support.
> Store-and-forward messaging dan background sync engine.

**Estimasi:** 4 hari
**Dependency:** Phase 06 (Mobile Shell), Phase 09 (Real-time)
**Output:** Offline-capable app dengan reliable sync.

---

## Task 19.1: Local Database Setup

**Input:** Phase 06 app shell
**Output:** SQLite/WatermelonDB local database

### Steps:
1. Setup WatermelonDB:
   ```typescript
   // src/database/index.ts
   import { Database } from '@nozbe/watermelondb';
   import SQLiteAdapter from '@nozbe/watermelondb/adapters/sqlite';
   import { schema } from './schema';
   import { migrations } from './migrations';

   // Models
   import { MessageModel } from './models/Message';
   import { ChatModel } from './models/Chat';
   import { ContactModel } from './models/Contact';
   import { DocumentModel } from './models/Document';

   const adapter = new SQLiteAdapter({
     schema,
     migrations,
     jsi: true, // Use JSI for performance
     onSetUpError: (error) => {
       console.error('Database setup error:', error);
     },
   });

   export const database = new Database({
     adapter,
     modelClasses: [
       MessageModel,
       ChatModel,
       ContactModel,
       DocumentModel,
     ],
   });
   ```
2. Define local schema:
   ```typescript
   // src/database/schema.ts
   import { appSchema, tableSchema } from '@nozbe/watermelondb';

   export const schema = appSchema({
     version: 1,
     tables: [
       tableSchema({
         name: 'chats',
         columns: [
           { name: 'server_id', type: 'string', isIndexed: true },
           { name: 'type', type: 'string' }, // personal, group
           { name: 'name', type: 'string', isOptional: true },
           { name: 'last_message', type: 'string', isOptional: true },
           { name: 'last_message_at', type: 'number', isOptional: true },
           { name: 'unread_count', type: 'number' },
           { name: 'is_muted', type: 'boolean' },
           { name: 'synced_at', type: 'number' },
         ],
       }),
       tableSchema({
         name: 'messages',
         columns: [
           { name: 'server_id', type: 'string', isIndexed: true },
           { name: 'chat_id', type: 'string', isIndexed: true },
           { name: 'sender_id', type: 'string' },
           { name: 'sender_name', type: 'string' },
           { name: 'content', type: 'string', isOptional: true },
           { name: 'type', type: 'string' }, // text, image, file, document_card
           { name: 'status', type: 'string' }, // sending, sent, delivered, read
           { name: 'reply_to_id', type: 'string', isOptional: true },
           { name: 'metadata', type: 'string', isOptional: true }, // JSON
           { name: 'created_at', type: 'number' },
           { name: 'is_pending', type: 'boolean' }, // true for unsent messages
         ],
       }),
       tableSchema({
         name: 'contacts',
         columns: [
           { name: 'server_id', type: 'string', isIndexed: true },
           { name: 'name', type: 'string' },
           { name: 'phone', type: 'string' },
           { name: 'avatar_url', type: 'string', isOptional: true },
           { name: 'status_text', type: 'string', isOptional: true },
           { name: 'is_registered', type: 'boolean' },
           { name: 'synced_at', type: 'number' },
         ],
       }),
       tableSchema({
         name: 'documents',
         columns: [
           { name: 'server_id', type: 'string', isIndexed: true },
           { name: 'title', type: 'string' },
           { name: 'icon', type: 'string', isOptional: true },
           { name: 'locked', type: 'boolean' },
           { name: 'owner_name', type: 'string' },
           { name: 'context_type', type: 'string' },
           { name: 'context_id', type: 'string', isOptional: true },
           { name: 'updated_at', type: 'number' },
           { name: 'synced_at', type: 'number' },
         ],
       }),
     ],
   });
   ```
3. Model classes:
   ```typescript
   // src/database/models/Message.ts
   import { Model } from '@nozbe/watermelondb';
   import { field, text, date, readonly, json } from '@nozbe/watermelondb/decorators';

   export class MessageModel extends Model {
     static table = 'messages';

     @text('server_id') serverId!: string;
     @text('chat_id') chatId!: string;
     @text('sender_id') senderId!: string;
     @text('sender_name') senderName!: string;
     @text('content') content!: string;
     @text('type') type!: string;
     @text('status') status!: string;
     @text('reply_to_id') replyToId!: string;
     @text('metadata') metadata!: string;
     @field('created_at') createdAt!: number;
     @field('is_pending') isPending!: boolean;
   }
   ```

### Acceptance Criteria:
- [ ] WatermelonDB initialized with JSI
- [ ] Schema: chats, messages, contacts, documents
- [ ] Models with proper decorators
- [ ] Indexes on server_id fields
- [ ] Database singleton exported

### Testing:
- [ ] Unit test: database initialization
- [ ] Unit test: insert/query/update/delete operations
- [ ] Unit test: indexes improve query speed

---

## Task 19.2: Store-and-Forward Messaging

**Input:** Task 19.1, Phase 09 (Real-time)
**Output:** Offline message queue dengan retry

### Steps:
1. Buat `src/services/MessageQueue.ts`:
   ```typescript
   class MessageQueue {
     private queue: PendingMessage[] = [];
     private isProcessing = false;

     async enqueue(message: PendingMessage): Promise<void> {
       // 1. Save to local DB with isPending = true, status = 'sending'
       await database.write(async () => {
         await database.get<MessageModel>('messages').create((m) => {
           m.chatId = message.chatId;
           m.senderId = message.senderId;
           m.content = message.content;
           m.type = message.type;
           m.status = 'sending';
           m.isPending = true;
           m.createdAt = Date.now();
         });
       });

       // 2. Add to queue
       this.queue.push(message);

       // 3. Process queue
       this.processQueue();
     }

     private async processQueue(): Promise<void> {
       if (this.isProcessing) return;
       this.isProcessing = true;

       while (this.queue.length > 0) {
         const msg = this.queue[0];

         try {
           // Send via API or WebSocket
           const result = await this.send(msg);

           // Update local DB: isPending = false, status = 'sent', server_id = result.id
           await this.markSent(msg.localId, result.id);

           // Remove from queue
           this.queue.shift();
         } catch (error) {
           if (this.isRetryable(error)) {
             // Wait and retry
             await this.delay(this.getBackoff(msg.retryCount));
             msg.retryCount++;
           } else {
             // Mark failed
             await this.markFailed(msg.localId);
             this.queue.shift();
           }
         }
       }

       this.isProcessing = false;
     }

     private getBackoff(retryCount: number): number {
       return Math.min(1000 * Math.pow(2, retryCount), 30000); // max 30s
     }
   }
   ```
2. Message status flow:
   ```
   sending → sent → delivered → read
      ↓
   failed (tap to retry)
   ```
3. UI indicators:
   - Sending: clock icon (⏳)
   - Sent: single check (✓)
   - Delivered: double check (✓✓)
   - Read: blue double check (✓✓)
   - Failed: red (!) with "Tap to retry"
4. Offline detection:
   ```typescript
   import NetInfo from '@react-native-community/netinfo';

   NetInfo.addEventListener((state) => {
     if (state.isConnected) {
       messageQueue.processQueue(); // flush pending messages
     }
   });
   ```

### Acceptance Criteria:
- [ ] Messages saved locally immediately (optimistic)
- [ ] Offline messages queued and sent when online
- [ ] Retry with exponential backoff
- [ ] Status indicators: sending, sent, delivered, read, failed
- [ ] Failed messages: tap to retry
- [ ] Come online → auto-flush queue
- [ ] Message order preserved

### Testing:
- [ ] Unit test: enqueue + process
- [ ] Unit test: retry logic (backoff)
- [ ] Unit test: mark sent/failed
- [ ] Unit test: online recovery
- [ ] Component test: status indicators render correctly

---

## Task 19.3: Sync Engine

**Input:** Task 19.1, all backend APIs
**Output:** Background sync for chats, messages, contacts, documents

### Steps:
1. Buat `src/services/SyncEngine.ts`:
   ```typescript
   class SyncEngine {
     private lastSyncTimestamp: Record<string, number> = {};

     async fullSync(): Promise<void> {
       console.log('[Sync] Starting full sync...');
       await Promise.all([
         this.syncChats(),
         this.syncContacts(),
       ]);
       console.log('[Sync] Full sync complete');
     }

     async syncChats(): Promise<void> {
       const lastSync = this.lastSyncTimestamp['chats'] || 0;
       const serverChats = await api.getChats({ updatedSince: lastSync });

       await database.write(async () => {
         for (const chat of serverChats) {
           const existing = await this.findByServerId('chats', chat.id);
           if (existing) {
             await existing.update((c: ChatModel) => {
               c.name = chat.name;
               c.lastMessage = chat.lastMessage;
               c.lastMessageAt = new Date(chat.lastMessageAt).getTime();
               c.unreadCount = chat.unreadCount;
               c.syncedAt = Date.now();
             });
           } else {
             await database.get<ChatModel>('chats').create((c) => {
               c.serverId = chat.id;
               c.type = chat.type;
               c.name = chat.name;
               c.lastMessage = chat.lastMessage;
               c.lastMessageAt = new Date(chat.lastMessageAt).getTime();
               c.unreadCount = chat.unreadCount;
               c.isMuted = false;
               c.syncedAt = Date.now();
             });
           }
         }
       });

       this.lastSyncTimestamp['chats'] = Date.now();
     }

     async syncMessages(chatId: string): Promise<void> {
       // Fetch messages newer than last synced message
       const lastMessage = await this.getLastSyncedMessage(chatId);
       const cursor = lastMessage?.serverId;
       const serverMessages = await api.getMessages(chatId, { after: cursor, limit: 50 });

       await database.write(async () => {
         for (const msg of serverMessages) {
           // Insert if not exists
           const exists = await this.findByServerId('messages', msg.id);
           if (!exists) {
             await database.get<MessageModel>('messages').create((m) => {
               m.serverId = msg.id;
               m.chatId = chatId;
               m.senderId = msg.senderId;
               m.senderName = msg.senderName;
               m.content = msg.content;
               m.type = msg.type;
               m.status = msg.status;
               m.isPending = false;
               m.createdAt = new Date(msg.createdAt).getTime();
             });
           }
         }
       });
     }

     async syncContacts(): Promise<void> {
       // Similar pattern to syncChats
     }
   }
   ```
2. Sync triggers:
   - App launch → full sync
   - App foreground → incremental sync
   - Pull-to-refresh → force sync
   - WebSocket reconnect → sync since last event
3. Conflict resolution:
   - Server wins for remote data
   - Local pending messages preserved
   - Merge strategy: update local with server data, keep pending
4. Sync status indicator:
   ```typescript
   // Show subtle sync indicator in header
   const SyncIndicator: React.FC = () => {
     const isSyncing = useSyncStore((s) => s.isSyncing);
     if (!isSyncing) return null;
     return <ActivityIndicator size="small" color="#6EE7B7" />;
   };
   ```

### Acceptance Criteria:
- [ ] Full sync on app launch
- [ ] Incremental sync on foreground
- [ ] Pull-to-refresh sync
- [ ] Server-wins conflict resolution
- [ ] Pending messages preserved during sync
- [ ] Sync indicator shown during sync
- [ ] Sync timestamp tracking

### Testing:
- [ ] Unit test: syncChats (insert + update)
- [ ] Unit test: syncMessages with cursor
- [ ] Unit test: conflict resolution
- [ ] Unit test: pending messages not overwritten
- [ ] Integration test: full sync flow

---

## Task 19.4: Offline UI Indicators

**Input:** Task 19.2, 19.3
**Output:** Clear offline indicators throughout app

### Steps:
1. Network status banner:
   ```typescript
   // src/components/NetworkBanner.tsx
   const NetworkBanner: React.FC = () => {
     const isConnected = useNetworkStore((s) => s.isConnected);

     if (isConnected) return null;

     return (
       <Animated.View style={styles.banner}>
         <WifiOffIcon size={16} color="#F59E0B" />
         <Text style={styles.text}>Tidak ada koneksi internet</Text>
       </Animated.View>
     );
   };
   ```
2. Chat list offline indicators:
   - Pending messages count badge
   - Last synced timestamp
3. Document offline handling:
   - Show cached document list
   - "Tersedia offline" badge for recently viewed docs
   - "Memerlukan koneksi" for unsynced docs
4. Graceful degradation:
   - Search: show local results + "Hasil lengkap memerlukan koneksi"
   - New chat: disabled offline
   - Existing chat: can compose + queue messages
   - Documents: view cached, queue edits if possible

### Acceptance Criteria:
- [ ] Network banner shown when offline
- [ ] Message queue visible (pending count)
- [ ] Cached data displayed from local DB
- [ ] Disabled actions clearly indicated
- [ ] Auto-recovery when online

### Testing:
- [ ] Component test: NetworkBanner renders when offline
- [ ] Component test: disabled actions have indicators
- [ ] Integration test: offline → compose → online → sent

---

## Phase 19 Review

### Testing Checklist:
- [ ] WatermelonDB: CRUD operations
- [ ] Message queue: enqueue, process, retry, fail
- [ ] Store-and-forward: offline → online delivery
- [ ] Sync engine: full + incremental
- [ ] Conflict resolution: server wins
- [ ] Offline indicators: banner + badges
- [ ] Cached data displayed offline
- [ ] Auto-recovery on reconnect

### Review Checklist:
- [ ] Local storage sesuai `spesifikasi-chatat.md` section 9.2
- [ ] Store-and-forward sesuai spec 9.2.1
- [ ] Message status sesuai spec 4.5 (✓/✓✓/blue)
- [ ] Indonesian labels on all indicators
- [ ] Performance: local queries < 50ms
- [ ] Commit: `feat(offline): implement local storage and sync engine`
