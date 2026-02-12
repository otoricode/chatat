# Phase 23: Performance Optimization

> Optimasi performa aplikasi mobile dan backend.
> Target: smooth 60fps, fast load, minimal memory.

**Estimasi:** 3 hari
**Dependency:** All feature phases (07-15)
**Output:** Optimized app ready for production loads.

---

## Task 23.1: Mobile Performance (React Native)

**Input:** All mobile screens
**Output:** Optimized rendering dan memory management

### Steps:
1. FlatList optimization untuk chat list dan messages:
   ```typescript
   // Chat message list
   <FlatList
     data={messages}
     renderItem={renderMessage}
     keyExtractor={(item) => item.id}
     // Performance props
     removeClippedSubviews={true}
     maxToRenderPerBatch={15}
     windowSize={10}
     initialNumToRender={20}
     getItemLayout={getMessageLayout} // if possible
     // Inverted for chat (newest at bottom)
     inverted
     // Prevent re-renders
     extraData={null}
   />

   // Memoize renderItem
   const renderMessage = useCallback(({ item }: { item: Message }) => (
     <MemoizedMessageBubble message={item} />
   ), []);

   // Memoize components
   const MemoizedMessageBubble = React.memo(MessageBubble, (prev, next) => {
     return prev.message.id === next.message.id
       && prev.message.status === next.message.status;
   });
   ```
2. Image optimization:
   ```typescript
   // Use react-native-fast-image for caching
   import FastImage from 'react-native-fast-image';

   <FastImage
     source={{
       uri: imageUrl,
       priority: FastImage.priority.normal,
       cache: FastImage.cacheControl.immutable,
     }}
     resizeMode={FastImage.resizeMode.cover}
     style={styles.image}
   />

   // Preload images for chat avatars
   const preloadAvatars = (contacts: Contact[]) => {
     FastImage.preload(
       contacts
         .filter((c) => c.avatarUrl)
         .map((c) => ({ uri: c.avatarUrl }))
     );
   };
   ```
3. Zustand store optimization:
   ```typescript
   // Use selectors to prevent unnecessary re-renders
   // BAD:
   const store = useChatStore();
   // GOOD:
   const messages = useChatStore((s) => s.messages);
   const unreadCount = useChatStore((s) => s.unreadCount);

   // Use shallow comparison for object selectors
   import { shallow } from 'zustand/shallow';
   const { messages, isLoading } = useChatStore(
     (s) => ({ messages: s.messages, isLoading: s.isLoading }),
     shallow
   );
   ```
4. Lazy loading screens:
   ```typescript
   // React.lazy for screens not immediately needed
   const DocumentEditor = React.lazy(() => import('./screens/DocumentEditor'));
   const Settings = React.lazy(() => import('./screens/Settings'));
   const Search = React.lazy(() => import('./screens/Search'));
   ```
5. Memory management:
   - Cleanup WebSocket listeners on unmount
   - Cancel API requests on navigation away
   - Release image cache when memory low
   - Limit messages in memory (virtualized list handles this)

### Acceptance Criteria:
- [ ] Chat list: smooth scroll 60fps
- [ ] Message list: smooth scroll with 1000+ messages
- [ ] Image caching with FastImage
- [ ] Zustand selectors prevent unnecessary re-renders
- [ ] Lazy loaded screens
- [ ] Memory stable (no leaks over time)

### Testing:
- [ ] Performance test: scroll FlatList with 1000 items
- [ ] Performance test: measure re-render count
- [ ] Performance test: memory usage over 10 min session
- [ ] Profile: React DevTools Profiler

---

## Task 23.2: Backend Performance (Go)

**Input:** All backend services
**Output:** Optimized queries, connection pooling, caching

### Steps:
1. Database connection pooling:
   ```go
   // pgx pool configuration
   poolConfig, _ := pgxpool.ParseConfig(dbURL)
   poolConfig.MaxConns = 25
   poolConfig.MinConns = 5
   poolConfig.MaxConnLifetime = time.Hour
   poolConfig.MaxConnIdleTime = 30 * time.Minute
   poolConfig.HealthCheckPeriod = time.Minute
   ```
2. Query optimization:
   ```go
   // Add missing indexes
   CREATE INDEX CONCURRENTLY idx_messages_chat_created
       ON messages(chat_id, created_at DESC);
   CREATE INDEX CONCURRENTLY idx_messages_sender
       ON messages(sender_id);
   CREATE INDEX CONCURRENTLY idx_chat_members_user
       ON chat_members(user_id);
   CREATE INDEX CONCURRENTLY idx_documents_context
       ON documents(chat_id, created_at DESC);
   CREATE INDEX CONCURRENTLY idx_topic_messages_topic
       ON topic_messages(topic_id, created_at DESC);
   CREATE INDEX CONCURRENTLY idx_document_entities_doc
       ON document_entities(document_id);
   CREATE INDEX CONCURRENTLY idx_document_entities_entity
       ON document_entities(entity_id);
   ```
3. Redis caching strategy:
   ```go
   // Cache frequently accessed data
   type CacheService struct {
       redis *redis.Client
   }

   // User profile: cache 5 min
   func (c *CacheService) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
       key := "user:" + userID.String()
       cached, err := c.redis.Get(ctx, key).Result()
       if err == nil {
           var user model.User
           json.Unmarshal([]byte(cached), &user)
           return &user, nil
       }
       return nil, err
   }

   func (c *CacheService) SetUser(ctx context.Context, user *model.User) error {
       data, _ := json.Marshal(user)
       return c.redis.Set(ctx, "user:"+user.ID.String(), data, 5*time.Minute).Err()
   }

   // Chat list: cache 1 min
   // Online status: cache 30s
   // Contact list: cache 5 min
   ```
4. API response compression:
   ```go
   import "github.com/go-chi/chi/v5/middleware"
   r.Use(middleware.Compress(5)) // gzip level 5
   ```
5. Pagination enforcement:
   ```go
   // All list endpoints: max 100 items per page
   func parsePagination(r *http.Request) (offset, limit int) {
       offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
       limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
       if limit <= 0 || limit > 100 {
           limit = 20
       }
       if offset < 0 {
           offset = 0
       }
       return
   }
   ```

### Acceptance Criteria:
- [ ] Connection pool: 25 max, 5 min connections
- [ ] All query indexes created
- [ ] Redis caching: user (5min), chat list (1min), online (30s)
- [ ] Response compression (gzip)
- [ ] Pagination max 100 items
- [ ] API response time < 200ms (p95)

### Testing:
- [ ] Benchmark: message query with/without index
- [ ] Benchmark: cached vs uncached user lookup
- [ ] Load test: 100 concurrent users
- [ ] Load test: 1000 messages/min throughput
- [ ] Monitor: connection pool utilization

---

## Task 23.3: Bundle Size & Startup

**Input:** React Native app
**Output:** Optimized bundle size and startup time

### Steps:
1. Bundle analysis:
   ```bash
   # Analyze bundle size
   npx react-native-bundle-visualizer

   # Target: < 10 MB (Android APK), < 15 MB (iOS IPA)
   ```
2. Tree shaking and dead code elimination:
   ```typescript
   // Import only what's needed
   // BAD:
   import _ from 'lodash';
   // GOOD:
   import debounce from 'lodash/debounce';

   // BAD:
   import { format, parse, add, sub, ... } from 'date-fns';
   // GOOD:
   import format from 'date-fns/format';
   ```
3. Hermes engine (Android):
   - Ensure Hermes enabled in android/app/build.gradle
   - Pre-compiled bytecode for faster startup
4. Splash screen optimization:
   ```typescript
   // Keep splash screen visible until app is ready
   import * as SplashScreen from 'expo-splash-screen';

   SplashScreen.preventAutoHideAsync();

   const App = () => {
     const [isReady, setIsReady] = useState(false);

     useEffect(() => {
       async function prepare() {
         // Load fonts
         await Font.loadAsync({ ... });
         // Restore auth state
         await authStore.restore();
         // Init database
         await database.initialize();
         setIsReady(true);
       }
       prepare();
     }, []);

     useEffect(() => {
       if (isReady) {
         SplashScreen.hideAsync();
       }
     }, [isReady]);
   };
   ```
5. Startup time targets:
   - Cold start: < 2 seconds
   - Warm start: < 500ms
   - TTI (time to interactive): < 3 seconds

### Acceptance Criteria:
- [ ] Android APK < 10 MB
- [ ] iOS IPA < 15 MB
- [ ] Cold start < 2 seconds
- [ ] Hermes enabled
- [ ] No unnecessary dependencies in bundle
- [ ] Splash screen until ready

### Testing:
- [ ] Measure: APK/IPA size
- [ ] Measure: cold start time
- [ ] Measure: TTI
- [ ] Bundle visualizer: no unexpected large deps

---

## Phase 23 Review

### Testing Checklist:
- [ ] Mobile: 60fps scroll in all lists
- [ ] Mobile: FastImage caching
- [ ] Mobile: no memory leaks
- [ ] Backend: query indexes effective
- [ ] Backend: Redis caching working
- [ ] Backend: < 200ms p95 response time
- [ ] Bundle: under size targets
- [ ] Startup: under time targets

### Review Checklist:
- [ ] Performance sesuai `spesifikasi-chatat.md` targets
- [ ] No regression in existing features
- [ ] Performance metrics documented
- [ ] Commit: `perf: optimize mobile rendering and backend queries`
