# Phase 17: Search System

> Implementasi full-text search untuk pesan, dokumen, dan kontak.
> Menggunakan PostgreSQL full-text search (tsvector/tsquery).

**Estimasi:** 3 hari
**Dependency:** Phase 07 (Chat), Phase 12 (Documents), Phase 15 (Entities)
**Output:** Global search + contextual search.

---

## Task 17.1: Search Backend (PostgreSQL FTS)

**Input:** Existing messages, documents, contacts tables
**Output:** Full-text search implementation

### Steps:
1. Add full-text search indexes:
   ```sql
   -- Migration: add_search_indexes
   -- Messages FTS
   ALTER TABLE messages ADD COLUMN search_vector tsvector;
   CREATE INDEX idx_messages_search ON messages USING GIN(search_vector);

   CREATE OR REPLACE FUNCTION messages_search_update() RETURNS trigger AS $$
   BEGIN
     NEW.search_vector := to_tsvector('indonesian', COALESCE(NEW.content, ''));
     RETURN NEW;
   END;
   $$ LANGUAGE plpgsql;

   CREATE TRIGGER messages_search_trigger
     BEFORE INSERT OR UPDATE ON messages
     FOR EACH ROW EXECUTE FUNCTION messages_search_update();

   -- Documents FTS
   ALTER TABLE documents ADD COLUMN search_vector tsvector;
   CREATE INDEX idx_documents_search ON documents USING GIN(search_vector);

   -- Blocks FTS (for content search within documents)
   ALTER TABLE blocks ADD COLUMN search_vector tsvector;
   CREATE INDEX idx_blocks_search ON blocks USING GIN(search_vector);

   -- Users FTS (name, phone, status)
   ALTER TABLE users ADD COLUMN search_vector tsvector;
   CREATE INDEX idx_users_search ON users USING GIN(search_vector);

   -- Entities FTS
   ALTER TABLE entities ADD COLUMN search_vector tsvector;
   CREATE INDEX idx_entities_search ON entities USING GIN(search_vector);
   ```
2. Buat `internal/service/search_service.go`:
   ```go
   type SearchService interface {
       SearchAll(ctx context.Context, userID uuid.UUID, query string, limit int) (*SearchResults, error)
       SearchMessages(ctx context.Context, userID uuid.UUID, query string, opts SearchOpts) (*MessageSearchResults, error)
       SearchDocuments(ctx context.Context, userID uuid.UUID, query string, opts SearchOpts) (*DocumentSearchResults, error)
       SearchContacts(ctx context.Context, userID uuid.UUID, query string) ([]*model.User, error)
       SearchEntities(ctx context.Context, userID uuid.UUID, query string) ([]*model.Entity, error)
       SearchInChat(ctx context.Context, chatID uuid.UUID, query string, opts SearchOpts) (*MessageSearchResults, error)
   }

   type SearchResults struct {
       Messages  []*MessageSearchResult  `json:"messages"`
       Documents []*DocumentSearchResult `json:"documents"`
       Contacts  []*model.User           `json:"contacts"`
       Entities  []*model.Entity         `json:"entities"`
   }

   type MessageSearchResult struct {
       Message    model.Message `json:"message"`
       ChatName   string        `json:"chatName"`
       SenderName string        `json:"senderName"`
       Highlight  string        `json:"highlight"` // snippet with <mark> tags
   }

   type DocumentSearchResult struct {
       Document  model.Document `json:"document"`
       Highlight string         `json:"highlight"` // matching block content
   }

   type SearchOpts struct {
       Offset int
       Limit  int
       ChatID *uuid.UUID // scope to specific chat
   }
   ```
3. Search query processing:
   ```go
   func buildTSQuery(query string) string {
       words := strings.Fields(query)
       escaped := make([]string, len(words))
       for i, w := range words {
           escaped[i] = w + ":*" // prefix matching
       }
       return strings.Join(escaped, " & ")
   }
   ```
4. Access control:
   - Messages: only from user's chats
   - Documents: only user's own, collaborator, or context member
   - Contacts: only user's synced contacts
   - Entities: only user's own entities
5. Highlight generation:
   ```go
   // Use ts_headline for snippet generation
   SELECT ts_headline('indonesian', content,
       plainto_tsquery('indonesian', $1),
       'MaxWords=30, MinWords=10, StartSel=<mark>, StopSel=</mark>'
   ) AS highlight
   ```

### Acceptance Criteria:
- [ ] Full-text search: messages, documents, contacts, entities
- [ ] Indonesian language support in tsvector
- [ ] Prefix matching (partial word search)
- [ ] Access control enforced
- [ ] Highlighted snippets returned
- [ ] Pagination support
- [ ] Chat-scoped message search

### Testing:
- [ ] Unit test: buildTSQuery
- [ ] Unit test: search messages (authorized chats only)
- [ ] Unit test: search documents (authorized only)
- [ ] Unit test: highlighted snippets
- [ ] Integration test: insert + search + find
- [ ] Integration test: access control

---

## Task 17.2: Search Handler & Endpoints

**Input:** Task 17.1
**Output:** REST endpoints untuk search

### Steps:
1. Buat `internal/handler/search_handler.go`:
   ```go
   // Global search (returns mixed results)
   GET /api/v1/search?q=query&limit=20

   // Type-specific search
   GET /api/v1/search/messages?q=query&offset=0&limit=20
   GET /api/v1/search/documents?q=query&offset=0&limit=20
   GET /api/v1/search/contacts?q=query
   GET /api/v1/search/entities?q=query

   // Chat-scoped search
   GET /api/v1/chats/:chatId/search?q=query&offset=0&limit=20
   ```
2. Global search: return top 3 of each type
3. Type-specific: full paginated results
4. Minimum query length: 2 characters

### Acceptance Criteria:
- [ ] Global search returns mixed results
- [ ] Type-specific search with pagination
- [ ] Chat-scoped search
- [ ] 2-char minimum query
- [ ] Empty results handled gracefully

### Testing:
- [ ] Integration test: global search returns mixed
- [ ] Integration test: paginated message search
- [ ] Integration test: chat-scoped search

---

## Task 17.3: Search UI (Mobile)

**Input:** Task 17.2, Phase 06 design system
**Output:** Search screen dan in-chat search

### Steps:
1. Global search screen:
   ```typescript
   // src/screens/SearchScreen.tsx
   const SearchScreen: React.FC = () => {
     const [query, setQuery] = useState('');
     const [activeTab, setActiveTab] = useState<'all' | 'messages' | 'documents' | 'contacts' | 'entities'>('all');

     return (
       <View style={styles.container}>
         <SearchBar
           value={query}
           onChangeText={setQuery}
           placeholder="Cari pesan, dokumen, kontak..."
           autoFocus
         />
         <TabBar
           tabs={[
             { key: 'all', label: 'Semua' },
             { key: 'messages', label: 'Pesan' },
             { key: 'documents', label: 'Dokumen' },
             { key: 'contacts', label: 'Kontak' },
             { key: 'entities', label: 'Entity' },
           ]}
           activeTab={activeTab}
           onTabChange={setActiveTab}
         />
         {activeTab === 'all' && <SearchAllResults query={query} />}
         {activeTab === 'messages' && <SearchMessageResults query={query} />}
         {activeTab === 'documents' && <SearchDocumentResults query={query} />}
         {/* ... */}
       </View>
     );
   };
   ```
2. Search result components:
   ```typescript
   // MessageSearchResultItem
   const MessageSearchResultItem: React.FC<{result: MessageSearchResult}> = ({ result }) => (
     <TouchableOpacity
       style={styles.resultItem}
       onPress={() => navigateToMessage(result.message)}
     >
       <Avatar name={result.senderName} size={40} />
       <View style={styles.resultContent}>
         <Text style={styles.chatName}>{result.chatName}</Text>
         <Text style={styles.senderName}>{result.senderName}</Text>
         <HighlightedText text={result.highlight} />
         <Text style={styles.date}>{formatDate(result.message.createdAt)}</Text>
       </View>
     </TouchableOpacity>
   );

   // HighlightedText: render <mark> as green highlighted spans
   const HighlightedText: React.FC<{text: string}> = ({ text }) => {
     const parts = text.split(/(<mark>.*?<\/mark>)/g);
     return (
       <Text style={styles.highlightText}>
         {parts.map((part, i) =>
           part.startsWith('<mark>') ? (
             <Text key={i} style={styles.marked}>
               {part.replace(/<\/?mark>/g, '')}
             </Text>
           ) : (
             <Text key={i}>{part}</Text>
           )
         )}
       </Text>
     );
   };
   ```
3. In-chat search:
   - Search icon in chat header
   - Slide-down search bar
   - Results: scroll to message with highlight
   - Up/down navigation between results
4. Debounced search (300ms)
5. Empty state: "Ketik untuk mencari..."
6. No results state: "Tidak ditemukan hasil untuk '[query]'"

### Acceptance Criteria:
- [ ] Global search screen with tabs
- [ ] Highlighted search results
- [ ] Tap result → navigate to source
- [ ] In-chat search: find messages within chat
- [ ] Debounced input (300ms)
- [ ] Empty + no results states
- [ ] Smooth tab switching

### Testing:
- [ ] Component test: SearchScreen renders
- [ ] Component test: HighlightedText renders marks
- [ ] Component test: each result type renders
- [ ] Component test: in-chat search
- [ ] Component test: empty/no-results states

---

## Phase 17 Review

### Testing Checklist:
- [ ] PostgreSQL FTS indexes created
- [ ] Search: messages, documents, contacts, entities
- [ ] Indonesian language tsvector
- [ ] Prefix matching works
- [ ] Access control enforced
- [ ] Highlighted snippets
- [ ] Global search UI
- [ ] In-chat search
- [ ] Deep link from result to source
- [ ] `go test ./...` pass

### Review Checklist:
- [ ] Search sesuai `spesifikasi-chatat.md` section 7.4
- [ ] Indonesian labels
- [ ] Dark theme consistent
- [ ] Performance: search < 500ms
- [ ] Commit: `feat(search): implement full-text search system`
