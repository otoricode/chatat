# Phase 12: Document Data Layer

> Implementasi backend untuk dokumen kolaboratif: CRUD, blocks, collaborators.
> Phase ini menghasilkan API lengkap untuk document management.

**Estimasi:** 3 hari
**Dependency:** Phase 02 (Database), Phase 03 (API)
**Output:** Document CRUD API, block operations, collaborator management.

---

## Task 12.1: Document Service (Backend)

**Input:** Document repository dari Phase 02
**Output:** Business logic untuk document management

### Steps:
1. Buat `internal/service/document_service.go`:
   ```go
   type DocumentService interface {
       Create(ctx context.Context, input CreateDocumentInput) (*DocumentFull, error)
       GetByID(ctx context.Context, docID, userID uuid.UUID) (*DocumentFull, error)
       ListByContext(ctx context.Context, contextType string, contextID uuid.UUID) ([]*DocumentListItem, error)
       ListByUser(ctx context.Context, userID uuid.UUID, filters DocumentFilters) ([]*DocumentListItem, error)
       ListAll(ctx context.Context, userID uuid.UUID, filters DocumentFilters) ([]*DocumentListItem, error)
       Update(ctx context.Context, docID uuid.UUID, input UpdateDocumentInput) (*DocumentFull, error)
       Delete(ctx context.Context, docID, userID uuid.UUID) error
       Duplicate(ctx context.Context, docID, userID uuid.UUID) (*DocumentFull, error)
   }

   type CreateDocumentInput struct {
       Title       string    `json:"title" validate:"required,min=1,max=200"`
       Icon        string    `json:"icon"`
       Cover       string    `json:"cover"`
       ChatID      *uuid.UUID `json:"chatId"`
       TopicID     *uuid.UUID `json:"topicId"`
       IsStandalone bool     `json:"isStandalone"`
       TemplateID  string    `json:"templateId"` // optional
   }

   type DocumentFull struct {
       Document     model.Document              `json:"document"`
       Blocks       []*model.Block              `json:"blocks"`
       Collaborators []*DocumentCollaboratorInfo `json:"collaborators"`
       Signers      []*DocumentSignerInfo       `json:"signers"`
       Tags         []string                    `json:"tags"`
       Entities     []*model.Entity             `json:"entities"`
       History      []*model.DocumentHistory     `json:"history"`
   }

   type DocumentListItem struct {
       ID          uuid.UUID `json:"id"`
       Title       string    `json:"title"`
       Icon        string    `json:"icon"`
       Locked      bool      `json:"locked"`
       RequireSigs bool      `json:"requireSigs"`
       SignedCount int       `json:"signedCount"`
       TotalSigners int      `json:"totalSigners"`
       OwnerName   string    `json:"ownerName"`
       UpdatedAt   time.Time `json:"updatedAt"`
       ContextType string    `json:"contextType"` // chat, topic, standalone
   }

   type DocumentFilters struct {
       Status   string // "all", "draft", "locked", "pending_signature"
       Tag      string
       EntityID *uuid.UUID
       Query    string // search title
   }
   ```
2. Implementasi Create:
   - Validate context (chatID or topicID or standalone)
   - Validate user is member of context (if not standalone)
   - Create document + set owner
   - If templateID provided → create blocks from template
   - Log history: "Dibuat oleh [Nama]"
   - If in chat/topic → send document_card message (inline card)
3. Implementasi access control:
   - Owner: full access
   - Editor: edit content, no lock/delete
   - Viewer: read only
   - Context-based: all chat/topic members can view
4. Implementasi Delete:
   - Only owner can delete
   - Locked documents cannot be deleted
   - CASCADE: blocks, collaborators, signers, tags, entities

### Acceptance Criteria:
- [x] Document created in chat, topic, or standalone context
- [x] Template applied if selected
- [x] Access control enforced (owner, editor, viewer)
- [x] Context members have access
- [x] Locked documents immutable
- [x] Inline card message sent on creation
- [x] History logged

### Testing:
- [x] Unit test: create document (all contexts)
- [x] Unit test: create from template
- [x] Unit test: access control (owner/editor/viewer)
- [x] Unit test: context member access
- [x] Unit test: locked document immutable
- [x] Unit test: delete (allowed/locked)
- [x] Unit test: history logging

---

## Task 12.2: Block Service (Backend)

**Input:** Block repository dari Phase 02
**Output:** Block operations for document editing

### Steps:
1. Buat `internal/service/block_service.go`:
   ```go
   type BlockService interface {
       AddBlock(ctx context.Context, docID uuid.UUID, input AddBlockInput) (*model.Block, error)
       UpdateBlock(ctx context.Context, blockID uuid.UUID, input UpdateBlockInput) (*model.Block, error)
       DeleteBlock(ctx context.Context, blockID uuid.UUID) error
       MoveBlock(ctx context.Context, docID uuid.UUID, blockID uuid.UUID, newPosition int) error
       GetBlocks(ctx context.Context, docID uuid.UUID) ([]*model.Block, error)
       BatchUpdate(ctx context.Context, docID uuid.UUID, operations []BlockOperation) error
   }

   type AddBlockInput struct {
       Type       string          `json:"type" validate:"required"`
       Content    *string         `json:"content"`
       Position   int             `json:"position"` // sort_order
       Checked    *bool           `json:"checked"`
       Rows       *json.RawMessage `json:"rows"`
       Columns    *json.RawMessage `json:"columns"`
       Language   *string         `json:"language"`
       Emoji      *string         `json:"emoji"`
       Color      *string         `json:"color"`
       ParentID   *uuid.UUID      `json:"parentId"` // for toggle children
   }

   type BlockOperation struct {
       Type    string          `json:"type"` // "add", "update", "delete", "move"
       BlockID *uuid.UUID      `json:"blockId"`
       Data    json.RawMessage `json:"data"`
   }
   ```
2. Implementasi:
   - AddBlock: validate type, insert at position, shift others
   - UpdateBlock: check doc not locked, update fields
   - DeleteBlock: remove + cascade children (for toggle blocks)
   - MoveBlock: update sort_order, reorder affected blocks
   - BatchUpdate: atomic transaction for multiple operations
3. Block validation per type:
   - paragraph/heading: content required
   - checklist: content + checked required
   - table: rows + columns required
   - code: content + language
   - callout: content + emoji + color
   - divider: no content needed
4. Log all changes to document_history

### Acceptance Criteria:
- [x] All 13 block types supported
- [x] Position-based ordering (sort_order)
- [x] Move block: reorder correctly
- [x] Batch operations: atomic
- [x] Locked document: reject all writes
- [x] Toggle children: nested blocks
- [x] History logged for edits

### Testing:
- [x] Unit test: add each block type
- [x] Unit test: update block
- [x] Unit test: delete block (with children)
- [x] Unit test: move block (reorder)
- [x] Unit test: batch operations
- [x] Unit test: locked document rejection
- [x] Unit test: invalid block type

---

## Task 12.3: Collaborator & Template Service

**Input:** Task 12.1
**Output:** Collaborator management dan document templates

### Steps:
1. Extend DocumentService:
   ```go
   // Collaborator management
   AddCollaborator(ctx context.Context, docID, userID uuid.UUID, role string) error
   RemoveCollaborator(ctx context.Context, docID, userID uuid.UUID) error
   UpdateCollaboratorRole(ctx context.Context, docID, userID uuid.UUID, role string) error
   ```
2. Buat `internal/service/template_service.go`:
   ```go
   type TemplateService interface {
       GetTemplates() []*DocumentTemplate
       GetTemplate(id string) *DocumentTemplate
       ApplyTemplate(docID uuid.UUID, templateID string) ([]*model.Block, error)
   }

   type DocumentTemplate struct {
       ID     string           `json:"id"`
       Name   string           `json:"name"`
       Blocks []TemplateBlock  `json:"blocks"`
   }
   ```
3. Implementasi 8 templates sesuai spec 5.11:
   - **kosong**: title only
   - **notulen-rapat**: headings (Agenda, Peserta, Pembahasan, Keputusan)
   - **daftar-belanja**: table (Nama Barang, Jumlah, Harga Satuan, Total)
   - **catatan-keuangan**: table (Tanggal, Keterangan, Pemasukan, Pengeluaran, Saldo)
   - **catatan-kesehatan**: headings (Keluhan, Diagnosis, Obat, Dokter, Kunjungan Berikutnya)
   - **kesepakatan-bersama**: headings (Pihak, Isi Kesepakatan, Ketentuan, Area Tanda Tangan)
   - **catatan-pertanian**: table (Lahan, Tanaman, Tanggal Tanam, Hasil Panen, Catatan)
   - **inventaris-aset**: table (Nama Aset, Jenis, Lokasi, Kondisi, Catatan)
4. Templates stored as JSON in-memory (embedded)

### Acceptance Criteria:
- [x] Add/remove/update collaborator roles
- [x] Only owner can manage collaborators
- [x] 8 templates available
- [x] Template creates correct blocks
- [x] Template tables have correct columns

### Testing:
- [x] Unit test: add/remove/update collaborator
- [x] Unit test: non-owner cannot manage collaborators
- [x] Unit test: each template produces correct blocks
- [x] Unit test: template table column types

---

## Task 12.4: Document Handler & Endpoints

**Input:** Task 12.1, 12.2, 12.3
**Output:** REST endpoints untuk documents

### Steps:
1. Buat `internal/handler/document_handler.go`:
   - `POST /api/v1/documents` → create document
   - `GET /api/v1/documents/:docId` → get document with blocks
   - `PUT /api/v1/documents/:docId` → update document metadata
   - `DELETE /api/v1/documents/:docId` → delete document
   - `GET /api/v1/documents` → list all user's documents (with filters)
   - `GET /api/v1/chats/:chatId/documents` → list documents in chat
   - `GET /api/v1/topics/:topicId/documents` → list documents in topic
   - `GET /api/v1/templates` → list available templates
2. Block endpoints:
   - `POST /api/v1/documents/:docId/blocks` → add block
   - `PUT /api/v1/documents/:docId/blocks/:blockId` → update block
   - `DELETE /api/v1/documents/:docId/blocks/:blockId` → delete block
   - `PUT /api/v1/documents/:docId/blocks/reorder` → reorder blocks
   - `POST /api/v1/documents/:docId/blocks/batch` → batch operations
3. Collaborator endpoints:
   - `POST /api/v1/documents/:docId/collaborators` → add
   - `DELETE /api/v1/documents/:docId/collaborators/:userId` → remove
   - `PUT /api/v1/documents/:docId/collaborators/:userId` → update role
4. Tag endpoints:
   - `POST /api/v1/documents/:docId/tags` → add tag
   - `DELETE /api/v1/documents/:docId/tags/:tag` → remove tag
5. History endpoint:
   - `GET /api/v1/documents/:docId/history` → get edit history

### Acceptance Criteria:
- [x] All CRUD endpoints functioning
- [x] Document filters: status, tag, entity, search
- [x] Block operations: add, update, delete, reorder, batch
- [x] Collaborator management endpoints
- [x] Tag management endpoints
- [x] History endpoint
- [x] Authorization enforced on all endpoints

### Testing:
- [x] Integration test: create document → add blocks → read
- [x] Integration test: collaborator management
- [x] Integration test: document in chat context
- [x] Integration test: filters and search

---

## Phase 12 Review

### Testing Checklist:
- [x] Document CRUD: create, read, update, delete
- [x] Block operations: all 13 types
- [x] Templates: all 8 produce correct output
- [x] Collaborators: add, remove, role update
- [x] Tags: add, remove, filter
- [x] Access control: owner/editor/viewer enforced
- [x] Context access: chat/topic member access
- [x] Locked: immutable
- [x] History: all actions logged
- [x] `go test ./...` pass

### Review Checklist:
- [x] Documents sesuai `spesifikasi-chatat.md` section 5
- [x] Block types sesuai spec 5.3
- [x] Templates sesuai spec 5.11
- [x] Collaborator roles sesuai spec 5.7
- [x] Data structures sesuai spec 8.2.4, 8.2.5
- [x] Commit: `feat(doc): implement document data layer and API`
