# Phase 15: Entity System

> Implementasi entity (tag terstruktur) yang dapat di-tagging ke dokumen.
> Entity memungkinkan relasi data antar dokumen untuk filtering dan navigasi.

**Estimasi:** 3 hari
**Dependency:** Phase 05 (Contacts), Phase 12 (Document Data)
**Output:** Entity CRUD, entity-document linking, filtering, contact-as-entity.

---

## Task 15.1: Entity Service (Backend)

**Input:** Entity schema dari Phase 02
**Output:** Entity CRUD dan linking logic

### Steps:
1. Buat `internal/service/entity_service.go`:
   ```go
   type EntityService interface {
       Create(ctx context.Context, input CreateEntityInput) (*model.Entity, error)
       GetByID(ctx context.Context, entityID, userID uuid.UUID) (*model.Entity, error)
       List(ctx context.Context, userID uuid.UUID, filters EntityFilters) ([]*EntityListItem, error)
       Update(ctx context.Context, entityID uuid.UUID, input UpdateEntityInput) (*model.Entity, error)
       Delete(ctx context.Context, entityID, userID uuid.UUID) error
       Search(ctx context.Context, userID uuid.UUID, query string) ([]*model.Entity, error)

       // Linking
       LinkToDocument(ctx context.Context, entityID, docID uuid.UUID) error
       UnlinkFromDocument(ctx context.Context, entityID, docID uuid.UUID) error
       GetDocumentEntities(ctx context.Context, docID uuid.UUID) ([]*model.Entity, error)
       GetEntityDocuments(ctx context.Context, entityID uuid.UUID) ([]*DocumentListItem, error)

       // Contact-as-entity
       CreateFromContact(ctx context.Context, contactID, userID uuid.UUID) (*model.Entity, error)
   }

   type CreateEntityInput struct {
       Name   string            `json:"name" validate:"required,min=1,max=100"`
       Type   string            `json:"type" validate:"required"` // free-form type label
       Fields map[string]string `json:"fields"` // flexible key-value metadata
   }

   type EntityFilters struct {
       Type  string
       Query string
       DocID *uuid.UUID // entities linked to this document
   }

   type EntityListItem struct {
       model.Entity
       DocumentCount int `json:"documentCount"`
   }
   ```
2. Implementasi entity uniqueness:
   - Entity scoped per user (userID + name unique)
   - Case-insensitive name matching
   - Merging duplicate entities (sama nama + type)
3. Implementasi free-form typed entities:
   - Type is a simple string label (not enum)
   - Contoh type: "Orang", "Lahan", "Aset", "Proyek", "Lokasi"
   - User creates types as needed → auto-suggest from existing
4. Entity fields (flexible metadata):
   ```go
   // Stored as JSONB in PostgreSQL
   // Examples:
   // Orang: { "telepon": "08xx", "alamat": "..." }
   // Lahan: { "luas": "2 Ha", "lokasi": "Desa X" }
   // Aset: { "jenis": "Kendaraan", "kondisi": "Baik" }
   ```
5. Contact-as-entity:
   - Convert existing contact → entity
   - Pre-fill: name, phone, type = "Orang"
   - Two-way link maintained

### Acceptance Criteria:
- [ ] Entity CRUD: create, read, update, delete
- [ ] Free-form type (user-defined labels)
- [ ] Flexible fields (key-value JSONB)
- [ ] Entity link/unlink to documents
- [ ] Get entities by document
- [ ] Get documents by entity
- [ ] Contact-to-entity conversion
- [ ] Search entities by name
- [ ] Entity scoped per user

### Testing:
- [ ] Unit test: create entity (various types)
- [ ] Unit test: update entity fields
- [ ] Unit test: delete entity (cascade unlink)
- [ ] Unit test: link/unlink document
- [ ] Unit test: get entities by document
- [ ] Unit test: get documents by entity
- [ ] Unit test: contact-to-entity conversion
- [ ] Unit test: search by name (case-insensitive)
- [ ] Unit test: name uniqueness per user

---

## Task 15.2: Entity Handler & Endpoints

**Input:** Task 15.1
**Output:** REST endpoints untuk entities

### Steps:
1. Buat `internal/handler/entity_handler.go`:
   ```go
   // Entity CRUD
   POST   /api/v1/entities           → create entity
   GET    /api/v1/entities           → list user's entities (with filters)
   GET    /api/v1/entities/:id       → get entity detail
   PUT    /api/v1/entities/:id       → update entity
   DELETE /api/v1/entities/:id       → delete entity
   GET    /api/v1/entities/search    → search entities (query param: q)

   // Entity-Document linking
   POST   /api/v1/documents/:docId/entities         → link entity to doc
   DELETE /api/v1/documents/:docId/entities/:entityId → unlink
   GET    /api/v1/documents/:docId/entities          → get doc entities
   GET    /api/v1/entities/:id/documents              → get entity docs

   // Type suggestions
   GET    /api/v1/entities/types     → list distinct types for user (auto-suggest)

   // Contact-as-entity
   POST   /api/v1/entities/from-contact → create entity from existing contact
   ```
2. Response format:
   ```json
   {
     "id": "uuid",
     "name": "Pak Ahmad",
     "type": "Orang",
     "fields": {
       "telepon": "081234567890",
       "alamat": "Desa Sukamaju"
     },
     "documentCount": 3,
     "createdAt": "2025-01-01T00:00:00Z",
     "updatedAt": "2025-01-01T00:00:00Z"
   }
   ```
3. Pagination on entity list and entity documents

### Acceptance Criteria:
- [ ] All endpoints functioning
- [ ] Filters: type, search query
- [ ] Type suggestions endpoint
- [ ] Link/unlink working
- [ ] Contact-to-entity endpoint
- [ ] Pagination implemented

### Testing:
- [ ] Integration test: full entity CRUD
- [ ] Integration test: link entity → get from document
- [ ] Integration test: get documents from entity
- [ ] Integration test: type auto-suggest
- [ ] Integration test: contact-to-entity

---

## Task 15.3: Entity UI Components (Mobile)

**Input:** Task 15.2, Phase 06 design system
**Output:** Entity management screens dan inline tagging

### Steps:
1. Buat EntityListScreen:
   ```typescript
   // src/screens/EntityListScreen.tsx
   const EntityListScreen: React.FC = () => {
     // Top: type filter chips (horizontal scroll)
     // Search bar
     // FlatList of entities
     // FAB: "Buat Entity"
     // Each item: name, type badge, document count

     return (
       <View style={styles.container}>
         <SearchBar
           placeholder="Cari entity..."
           onSearch={setQuery}
         />
         <ScrollView horizontal style={styles.typeFilters}>
           <TypeChip label="Semua" active={!typeFilter} onPress={() => setTypeFilter(null)} />
           {types.map(type => (
             <TypeChip key={type} label={type} active={typeFilter === type} onPress={() => setTypeFilter(type)} />
           ))}
         </ScrollView>
         <FlatList
           data={filteredEntities}
           renderItem={({ item }) => (
             <EntityListItem
               entity={item}
               onPress={() => navigation.navigate('EntityDetail', { entityId: item.id })}
             />
           )}
         />
         <FAB icon="plus" onPress={showCreateEntitySheet} />
       </View>
     );
   };
   ```
2. Buat EntityDetailScreen:
   - Header: name, type badge
   - Fields section: key-value pairs
   - Linked documents: list with tap → open editor
   - Actions: Edit, Delete, "Buat dari Kontak"
3. Buat CreateEntitySheet (BottomSheet):
   ```typescript
   const CreateEntitySheet: React.FC = () => {
     return (
       <BottomSheet>
         <TextInput label="Nama" value={name} onChangeText={setName} />
         <TypeInput
           label="Tipe"
           value={type}
           onChangeText={setType}
           suggestions={existingTypes}
         />
         <DynamicFieldsInput
           fields={fields}
           onAdd={addField}
           onRemove={removeField}
           onUpdate={updateField}
         />
         <Button title="Simpan" onPress={saveEntity} />
       </BottomSheet>
     );
   };
   ```
4. Dynamic fields input:
   - "Tambah Field" button
   - Each field: key input + value input + remove button
   - Flexible: user defines their own fields

### Acceptance Criteria:
- [ ] Entity list with search and type filter
- [ ] Entity detail: fields + linked documents
- [ ] Create entity: name, type (with suggestions), dynamic fields
- [ ] Edit entity
- [ ] Delete entity with confirmation
- [ ] Type chips: horizontal scroll, active state
- [ ] FAB for quick creation

### Testing:
- [ ] Component test: EntityListScreen renders entities
- [ ] Component test: EntityDetailScreen fields + docs
- [ ] Component test: CreateEntitySheet validation
- [ ] Component test: type filter
- [ ] Component test: dynamic fields input

---

## Task 15.4: Entity Tagging in Documents

**Input:** Task 15.3, Phase 13 (Block Editor)
**Output:** Tag entity ke dokumen dari dalam editor

### Steps:
1. Entity tag bar di document editor:
   ```typescript
   // Shown below document title in editor
   const EntityTagBar: React.FC<{ docId: string }> = ({ docId }) => {
     const { data: entities } = useQuery(['docEntities', docId], () =>
       api.getDocumentEntities(docId)
     );

     return (
       <View style={styles.tagBar}>
         <ScrollView horizontal showsHorizontalScrollIndicator={false}>
           {entities?.map((entity) => (
             <Pressable
               key={entity.id}
               style={styles.tag}
               onPress={() => navigation.navigate('EntityDetail', { entityId: entity.id })}
               onLongPress={() => confirmUnlink(entity)}
             >
               <Text style={styles.tagType}>{entity.type}</Text>
               <Text style={styles.tagName}>{entity.name}</Text>
             </Pressable>
           ))}
           <TouchableOpacity
             style={styles.addTag}
             onPress={showEntitySelector}
           >
             <PlusIcon size={14} color="#6EE7B7" />
             <Text style={styles.addTagText}>Tag</Text>
           </TouchableOpacity>
         </ScrollView>
       </View>
     );
   };
   ```
2. Entity selector modal:
   - Search existing entities
   - Quick create new entity
   - Recent entities / suggested
   - Select → link → tag appears in bar
3. Tag pill styling:
   - Color-coded by type
   - Type prefix (small, dimmed) + Name (white)
   - Long press → unlink confirmation
   - Tap → navigate to entity detail
4. Entity filter in Dokumen tab:
   - Filter documents by linked entity
   - Show entity as filter chip
   - "Dokumen dengan [Entity Name]"

### Acceptance Criteria:
- [ ] Entity tag bar shown in document editor
- [ ] Add tag via selector (search + create)
- [ ] Remove tag via long press
- [ ] Tags color-coded by type
- [ ] Tap tag → entity detail
- [ ] Filter documents by entity in Dokumen tab
- [ ] Tag bar scrollable horizontally

### Testing:
- [ ] Component test: EntityTagBar renders tags
- [ ] Component test: add tag flow
- [ ] Component test: remove tag confirmation
- [ ] Component test: entity selector search + create
- [ ] Integration test: tag → filter works

---

## Phase 15 Review

### Testing Checklist:
- [ ] Entity CRUD: all operations
- [ ] Free-form types with auto-suggest
- [ ] Dynamic fields (key-value)
- [ ] Entity-document linking
- [ ] Contact-to-entity conversion
- [ ] Entity list: search + type filter
- [ ] Entity detail: fields + linked docs
- [ ] Tag bar in document editor
- [ ] Document filter by entity
- [ ] `go test ./...` pass

### Review Checklist:
- [ ] Entity model sesuai `spesifikasi-chatat.md` section 6
- [ ] Data structure sesuai spec 8.2.6
- [ ] Types sesuai spec 6.2 (Orang, Lahan, Aset, dll)
- [ ] Fields sesuai spec 6.3 (free-form)
- [ ] Tag UI sesuai spec 6.4
- [ ] Indonesian labels throughout
- [ ] Dark theme consistent
- [ ] Commit: `feat(entity): implement entity system with tagging`
