# Phase 14: Document Collaboration & Locking

> Implementasi real-time collaborative editing, locking system (manual + tanda tangan digital),
> dan integrasi dokumen ke dalam chat (inline card + tab Dokumen).

**Estimasi:** 5 hari
**Dependency:** Phase 09 (Real-time), Phase 12 (Document Data), Phase 13 (Block Editor)
**Output:** Kolaborasi real-time, locking system, digital signatures.

---

## Task 14.1: Real-time Document Sync

**Input:** Phase 09 WebSocket infra, Phase 12 document API
**Output:** Real-time update antar kolaborator saat editing bersamaan

### Steps:
1. Extend WebSocket hub untuk document channels:
   ```go
   // internal/websocket/document_channel.go
   type DocumentChannel struct {
       DocumentID uuid.UUID
       Clients    map[uuid.UUID]*Client
       mu         sync.RWMutex
   }

   type DocumentEvent struct {
       Type       string          `json:"type"` // block_update, block_add, block_delete, block_move, cursor_move
       DocumentID uuid.UUID       `json:"documentId"`
       BlockID    *uuid.UUID      `json:"blockId,omitempty"`
       UserID     uuid.UUID       `json:"userId"`
       UserName   string          `json:"userName"`
       Data       json.RawMessage `json:"data"`
       Timestamp  time.Time       `json:"timestamp"`
   }
   ```
2. Event types:
   - `doc:join` → user joins document editing session
   - `doc:leave` → user leaves
   - `doc:block_update` → block content changed
   - `doc:block_add` → new block added
   - `doc:block_delete` → block removed
   - `doc:block_move` → block reordered
   - `doc:cursor_move` → cursor position changed
   - `doc:lock_changed` → document locked/unlocked
3. Conflict resolution strategy (Operational Transform lite):
   ```go
   // Simple last-write-wins per block
   // Each block has version counter
   // Client sends: { blockId, version, content }
   // Server checks: if version matches → accept, broadcast
   //                 if version mismatch → reject, send current
   // Client re-applies on rejection
   ```
4. Implementasi cursor presence:
   - Show other users' cursor positions
   - Color-coded per user
   - Name label above cursor
   - Fade out after 5s idle

### Acceptance Criteria:
- [ ] Users join/leave document channel
- [ ] Block edits broadcast to other viewers in real-time
- [ ] Last-write-wins conflict resolution per block
- [ ] Cursor presence: colored cursors with names
- [ ] Smooth: no lag on typing (debounce sends 300ms)
- [ ] Reconnect: resync on WebSocket reconnect

### Testing:
- [ ] Unit test: document channel join/leave
- [ ] Unit test: broadcast event to all except sender
- [ ] Unit test: version conflict detection
- [ ] Integration test: 2 users editing different blocks
- [ ] Integration test: 2 users editing same block (conflict)
- [ ] Integration test: reconnect + resync

---

## Task 14.2: Document Locking System (Backend)

**Input:** Phase 12 document model
**Output:** Manual lock + signature-based lock

### Steps:
1. Extend document model:
   ```go
   // Lock modes
   const (
       LockNone      = "" // Draft - editable
       LockManual    = "manual" // Owner locked manually
       LockSignature = "signature" // Locked after all signatures
   )

   // Lock/Unlock endpoints
   // POST /api/v1/documents/:docId/lock
   type LockDocumentInput struct {
       Mode    string   `json:"mode" validate:"required,oneof=manual signature"`
       Message string   `json:"message"` // optional lock message
   }

   // POST /api/v1/documents/:docId/unlock
   // Only owner can unlock manual lock
   // Signature lock cannot be unlocked

   // Document signer management
   // POST /api/v1/documents/:docId/signers
   type AddSignerInput struct {
       UserID uuid.UUID `json:"userId" validate:"required"`
   }
   ```
2. Implementasi manual lock:
   - Owner calls lock endpoint
   - Document becomes read-only for all
   - Owner can unlock anytime
   - Lock message shown to collaborators
3. Implementasi signature-based lock:
   - Owner adds signers (from collaborators/chat members)
   - Owner requests signatures (status: `pending`)
   - Each signer can sign (status: `signed`)
   - When ALL signed → document auto-locks permanently
   - Signature lock cannot be unlocked
4. Sign endpoint:
   ```go
   // POST /api/v1/documents/:docId/sign
   type SignDocumentInput struct {
       PIN    string `json:"pin" validate:"required,len=6"`
   }

   // Server verifies PIN matches user's signing PIN
   // Stores: signer_id, signed_at, signature_hash (SHA-256 of doc content)
   ```
5. Lock status computation:
   ```go
   func ComputeLockStatus(doc *Document, signers []*DocumentSigner) string {
       if doc.Locked && doc.LockMode == LockManual {
           return "locked_manual" // icon: lock
       }
       if doc.LockMode == LockSignature {
           allSigned := true
           for _, s := range signers {
               if s.SignedAt == nil {
                   allSigned = false
                   break
               }
           }
           if allSigned {
               return "locked_signed" // icon: lock + checkmark
           }
           return "pending_signatures" // icon: pen
       }
       return "draft" // icon: file-edit
   }
   ```

### Acceptance Criteria:
- [ ] Manual lock: owner can lock/unlock
- [ ] Locked document: all edits rejected (400 Bad Request)
- [ ] Signature flow: add signers → request → sign → auto-lock
- [ ] PIN verification for signing
- [ ] Signature hash stored (SHA-256)
- [ ] Lock status correctly computed
- [ ] Only owner can manage signers
- [ ] WebSocket broadcasts lock state changes

### Testing:
- [ ] Unit test: manual lock/unlock
- [ ] Unit test: locked document rejects edits
- [ ] Unit test: add/remove signers
- [ ] Unit test: sign with valid PIN
- [ ] Unit test: sign with invalid PIN → rejected
- [ ] Unit test: all signed → auto-lock
- [ ] Unit test: signature lock cannot be unlocked
- [ ] Integration test: full signing flow

---

## Task 14.3: Lock & Signature UI (Mobile)

**Input:** Task 14.2, Phase 13 editor
**Output:** Lock controls dan signature flow di mobile

### Steps:
1. Lock status badge di document header:
   ```typescript
   const LockStatusBadge: React.FC<{ status: LockStatus }> = ({ status }) => {
     const config = {
       draft: { icon: 'file-edit', label: 'Draft', color: '#9CA3AF' },
       locked_manual: { icon: 'lock', label: 'Dikunci', color: '#F59E0B' },
       pending_signatures: { icon: 'pen-tool', label: 'Menunggu TTD', color: '#6EE7B7' },
       locked_signed: { icon: 'shield-check', label: 'Ditandatangani', color: '#10B981' },
     };
     // ... render badge
   };
   ```
2. Lock action sheet (owner only):
   - "Kunci Dokumen" → lock manually
   - "Minta Tanda Tangan" → open signer selector
   - "Buka Kunci" → unlock (only manual lock)
3. Signer management screen:
   ```typescript
   // SignerListScreen
   // - List signers with status (Menunggu / Sudah TTD)
   // - "Tambah Penanda Tangan" button
   // - Select from chat members
   // - "Minta Tanda Tangan" → sends notification + changes status
   ```
4. Sign confirmation modal:
   ```typescript
   // SignConfirmModal
   // - Shows document title, preview
   // - "Dengan menandatangani, Anda menyetujui isi dokumen ini"
   // - PIN input (6 digit)
   // - "Tanda Tangani" button
   // - Loading state
   // - Success animation (checkmark)
   ```
5. Document header actions:
   - Draft: "Kunci" button
   - Locked: "Terkunci" badge + "Buka Kunci" (if owner + manual)
   - Pending: "N/M sudah TTD" badge
   - Signed: "Ditandatangani" badge + "Lihat TTD" button

### Acceptance Criteria:
- [ ] Lock status badge shown on all documents
- [ ] Owner can lock/unlock via action sheet
- [ ] Signer selector from chat/group members
- [ ] PIN confirmation for signing
- [ ] Sign success animation
- [ ] Editor becomes read-only when locked
- [ ] Real-time lock status updates

### Testing:
- [ ] Component test: LockStatusBadge all states
- [ ] Component test: lock action sheet
- [ ] Component test: signer list
- [ ] Component test: sign confirmation modal
- [ ] Integration test: lock → editor read-only
- [ ] Integration test: full signing flow UI

---

## Task 14.4: Document in Chat (Inline Card)

**Input:** Phase 07 (Chat), Phase 08 (Group), Phase 12 (Documents)
**Output:** Inline document card di chat + Dokumen tab

### Steps:
1. Document card message component:
   ```typescript
   // src/components/chat/DocumentCardMessage.tsx
   const DocumentCardMessage: React.FC<Props> = ({ message }) => {
     const doc = message.metadata as DocumentCardData;

     return (
       <TouchableOpacity
         style={styles.card}
         onPress={() => navigation.navigate('DocumentEditor', { docId: doc.id })}
       >
         <View style={styles.cardHeader}>
           <Text style={styles.icon}>{doc.icon || '📄'}</Text>
           <View style={styles.cardInfo}>
             <Text style={styles.title}>{doc.title}</Text>
             <Text style={styles.meta}>{doc.ownerName} • {formatDate(doc.updatedAt)}</Text>
           </View>
           <LockStatusBadge status={doc.lockStatus} size="small" />
         </View>
         {doc.requireSigs && (
           <View style={styles.signatureBar}>
             <Text style={styles.sigText}>{doc.signedCount}/{doc.totalSigners} tanda tangan</Text>
             <View style={styles.sigProgress}>
               <View style={[styles.sigFill, { width: `${(doc.signedCount/doc.totalSigners)*100}%` }]} />
             </View>
           </View>
         )}
       </TouchableOpacity>
     );
   };
   ```
2. Card style (sesuai spec 5.12):
   - Background: #1F2937 (card bg di chat)
   - Border: 1px #374151
   - Rounded: 12px
   - Icon + title + status
   - Signature progress if applicable
3. Tab Dokumen di ChatScreen/GroupScreen:
   ```typescript
   // Tab in chat/group screen (swipeable)
   const DocumentsTab: React.FC<{ chatId: string }> = ({ chatId }) => {
     const { data: documents } = useQuery(['documents', chatId], () =>
       api.getDocumentsByChat(chatId)
     );

     return (
       <FlatList
         data={documents}
         renderItem={({ item }) => <DocumentListItem doc={item} />}
         ListHeaderComponent={
           <TouchableOpacity onPress={createDocument} style={styles.createButton}>
             <PlusIcon />
             <Text>Buat Dokumen</Text>
           </TouchableOpacity>
         }
         ListEmptyComponent={
           <EmptyState
             icon="file-text"
             title="Belum Ada Dokumen"
             subtitle="Buat dokumen pertama untuk mulai berkolaborasi"
           />
         }
       />
     );
   };
   ```
4. Create document flow:
   - FAB or "+" in Dokumen tab
   - Template selector
   - Create → opens editor
   - Auto-sends inline card to chat

### Acceptance Criteria:
- [ ] Document card renders in chat message bubble
- [ ] Tap card → opens document editor
- [ ] Lock status badge on card
- [ ] Signature progress bar when applicable
- [ ] Dokumen tab in chat/group screens
- [ ] Create document from chat → inline card sent
- [ ] Document list shows all docs in context

### Testing:
- [ ] Component test: DocumentCardMessage renders
- [ ] Component test: DocumentsTab empty + with data
- [ ] Component test: signature progress bar
- [ ] Integration test: create doc → card appears in chat
- [ ] Integration test: tap card → opens editor

---

## Task 14.5: Document History & Activity Log

**Input:** Phase 12 history model
**Output:** Activity log UI di dokumen

### Steps:
1. Backend: history service:
   ```go
   type HistoryEntry struct {
       ID         uuid.UUID `json:"id"`
       DocumentID uuid.UUID `json:"documentId"`
       UserID     uuid.UUID `json:"userId"`
       UserName   string    `json:"userName"`
       Action     string    `json:"action"` // created, edited, locked, unlocked, signed, collaborator_added, collaborator_removed
       Details    string    `json:"details"`
       CreatedAt  time.Time `json:"createdAt"`
   }

   // Auto-logged actions:
   // - "Dibuat oleh [Nama]"
   // - "[Nama] mengedit dokumen"
   // - "[Nama] mengunci dokumen"
   // - "[Nama] membuka kunci"
   // - "[Nama] menandatangani"
   // - "[Nama] ditambahkan sebagai [role]"
   // - "[Nama] dihapus dari kolaborator"
   ```
2. History UI (slide-up panel or modal):
   ```typescript
   const DocumentHistory: React.FC<{ docId: string }> = ({ docId }) => {
     const { data: history } = useQuery(['docHistory', docId], () =>
       api.getDocumentHistory(docId)
     );

     return (
       <FlatList
         data={history}
         renderItem={({ item }) => (
           <View style={styles.entry}>
             <Avatar user={item.userName} size={32} />
             <View style={styles.entryContent}>
               <Text style={styles.action}>{item.details}</Text>
               <Text style={styles.time}>{formatRelative(item.createdAt)}</Text>
             </View>
           </View>
         )}
       />
     );
   };
   ```

### Acceptance Criteria:
- [ ] All document actions logged
- [ ] History endpoint paginated
- [ ] History UI shows timeline
- [ ] Actions in Indonesian
- [ ] Avatar + name + timestamp

### Testing:
- [ ] Unit test: history logging for each action type
- [ ] Component test: history list renders
- [ ] Integration test: edit → history entry created

---

## Phase 14 Review

### Testing Checklist:
- [ ] Real-time sync: multi-user editing
- [ ] Conflict resolution: last-write-wins per block
- [ ] Manual lock: lock, unlock, reject edits
- [ ] Signature flow: add signers → request → sign → auto-lock
- [ ] PIN verification
- [ ] Lock status badge: all 4 states
- [ ] Inline card in chat
- [ ] Dokumen tab: list + create
- [ ] History: all actions logged + UI

### Review Checklist:
- [ ] Locking sesuai `spesifikasi-chatat.md` section 5.5, 5.6
- [ ] Signature flow sesuai spec 5.9
- [ ] Lock status badges sesuai spec 5.10
- [ ] Inline card sesuai spec 5.12
- [ ] Indonesian labels throughout
- [ ] Dark theme consistent
- [ ] Commit: `feat(collab): implement document collaboration and locking`
