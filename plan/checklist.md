# Checklist Audit

> Checklist lengkap untuk setiap phase. Gunakan sebagai referensi saat review.

---

## Cara Penggunaan

Setiap kali menyelesaikan satu phase, jalankan checklist ini:

1. Buka file phase yang bersangkutan
2. Pastikan semua acceptance criteria terpenuhi
3. Jalankan testing checklist
4. Jalankan review checklist
5. Centang di checklist ini
6. Commit dengan pesan yang sesuai

---

## Phase 01 — Project Setup
- [x] Go project initialized (`go mod init`)
- [x] React Native project initialized (CLI/Expo)
- [x] PostgreSQL + Redis Docker berjalan
- [x] Dev tools siap (golangci-lint, ESLint, Prettier)
- [x] Skeleton berjalan tanpa error

## Phase 02 — Database Layer
- [x] PostgreSQL database terbuat
- [x] Migration system berjalan (golang-migrate)
- [x] Semua tabel core terbuat: users, chats, messages, topics, documents, blocks, entities
- [x] Repository pattern untuk setiap tabel
- [x] `go test ./...` pass untuk semua repository

## Phase 03 — API & WebSocket Foundation
- [ ] HTTP router terkonfigurasi (Chi/Gin)
- [ ] Middleware stack: auth, CORS, rate limit, logging, recovery
- [ ] Error handling & response pattern
- [ ] WebSocket hub: connection, rooms, broadcast
- [ ] Health check endpoint berfungsi

## Phase 04 — Authentication System
- [ ] SMS OTP 6-digit: send, verify, expire
- [ ] Reverse OTP via WhatsApp: generate code, verify incoming WA message
- [ ] JWT token: generate, validate, refresh
- [ ] Phone number normalization (E.164)
- [ ] Session management: satu device per nomor

## Phase 05 — User & Contact System
- [ ] User profile CRUD (name, avatar emoji, status)
- [ ] Contact sync API: upload hashed phone numbers, match registered users
- [ ] Contact list endpoint + online/last seen status
- [ ] User search by phone number

## Phase 06 — Mobile App Shell
- [ ] React Navigation: stack + bottom tabs (Chat, Dokumen)
- [ ] Dark theme (WA-style colors dari spec)
- [ ] Typography setup (Plus Jakarta Sans, Inter, JetBrains Mono)
- [ ] Shared components: Loading, Empty, Error, Avatar, Badge
- [ ] FAB (Floating Action Button) per tab

## Phase 07 — Chat Personal
- [ ] Backend: personal chat creation, message send/receive API
- [ ] Backend: delivery status (sent, delivered, read)
- [ ] Frontend: Chat list dengan preview, badge unread, timestamp
- [ ] Frontend: Chat screen dengan message bubbles (kiri/kanan)
- [ ] Frontend: Reply, forward, delete message
- [ ] Frontend: Input bar dengan tombol kirim
- [ ] Tab Chat + Tab Dokumen dalam chat personal

## Phase 08 — Chat Group
- [ ] Backend: Group CRUD (create, update name/icon, delete)
- [ ] Backend: Member management (add, remove, admin promotion)
- [ ] Frontend: Group creation wizard (nama, ikon, pilih anggota)
- [ ] Frontend: Group chat screen + member list
- [ ] Frontend: Group settings (edit nama, ikon, anggota)
- [ ] Tab Chat + Tab Dokumen + Tab Topik dalam grup

## Phase 09 — Real-time Messaging
- [ ] WebSocket client di React Native
- [ ] Message relay server-to-client real-time
- [ ] Typing indicator ("sedang mengetik...")
- [ ] Online/offline status real-time
- [ ] Read receipts (✓ sent, ✓✓ delivered, biru = read)
- [ ] Auto-reconnect + offline queue

## Phase 10 — Topic System
- [ ] Backend: Topic CRUD (parent: personal/group)
- [ ] Backend: Topic membership (from parent members)
- [ ] Backend: Topic messages (same as chat messages)
- [ ] Frontend: Create topic dari chat/group
- [ ] Frontend: Topic list di tab Topik grup
- [ ] Frontend: Topic screen (Tab Diskusi + Tab Dokumen)

## Phase 11 — Media System
- [ ] Backend: File upload API (image, file)
- [ ] Backend: Image compression + thumbnail generation
- [ ] Backend: S3-compatible storage (MinIO/AWS S3)
- [ ] Frontend: Image picker + camera
- [ ] Frontend: Image preview + gallery view
- [ ] Frontend: File download + share

## Phase 12 — Document Data Layer
- [ ] Backend: Document CRUD API
- [ ] Backend: Block CRUD (add, update, delete, reorder)
- [ ] Backend: Document context (chatId/groupId/topicId/standalone)
- [ ] Backend: Collaborator management (owner, editor, viewer)
- [ ] Backend: Document history log

## Phase 13 — Block Editor
- [ ] 13 block types fully implemented
- [ ] Slash command menu (`/`)
- [ ] Floating toolbar (bold, italic, strikethrough, code, link)
- [ ] Table block (dynamic rows/columns, column types)
- [ ] Checklist block (interactive checkboxes)
- [ ] Toggle block (collapsible content)
- [ ] Template selection (8 templates)

## Phase 14 — Document Collaboration & Locking
- [ ] Real-time document sync via WebSocket
- [ ] Conflict resolution (OT atau CRDT)
- [ ] Manual lock by owner
- [ ] Signature-based lock (multi-signer)
- [ ] Signature flow UI (request, sign, lock)
- [ ] Lock status badges (Draft, Menunggu Tanda Tangan, Terkunci)
- [ ] Document inline card di chat + tab Dokumen

## Phase 15 — Entity System
- [ ] Backend: Entity CRUD (free-form tags)
- [ ] Backend: Entity-document linking (many-to-many)
- [ ] Backend: Contact-as-entity support
- [ ] Frontend: Entity picker/creator in document editor
- [ ] Frontend: Entity filter on document list
- [ ] Frontend: Entity search

## Phase 16 — Push Notifications
- [ ] FCM setup (Android)
- [ ] APNs setup (iOS)
- [ ] Notification types: new message, signature request, document shared
- [ ] Deep linking: tap notification → open relevant screen
- [ ] Notification preferences per chat/group

## Phase 17 — Search System
- [ ] Backend: Full-text search (messages, documents, contacts)
- [ ] Backend: Search indexing strategy
- [ ] Frontend: Global search bar
- [ ] Frontend: In-chat message search
- [ ] Frontend: Document search + entity filter

## Phase 18 — Internationalization
- [ ] i18n library setup (react-i18next / i18n-js)
- [ ] Bahasa Indonesia translations (default)
- [ ] English translations
- [ ] Arabic translations
- [ ] RTL layout support for Arabic
- [ ] Dynamic language switch

## Phase 19 — Local Storage & Sync
- [ ] SQLite/WatermelonDB local database setup
- [ ] Message store-and-forward (server relay → device)
- [ ] Offline message queue (send when back online)
- [ ] Sync engine: server timestamp comparison
- [ ] Chat history retained on device

## Phase 20 — Cloud Backup
- [ ] Google Drive backup (Android)
- [ ] iCloud backup (iOS)
- [ ] Backup flow UI (settings → backup → progress)
- [ ] Restore flow (fresh install → restore from cloud)
- [ ] Backup scheduling (manual / daily auto)

## Phase 21 — Settings & Preferences
- [ ] Settings screen layout
- [ ] Profile edit (nama, avatar, status)
- [ ] Language switch (ID/EN/AR)
- [ ] Notification preferences
- [ ] Storage & data management
- [ ] About + version info
- [ ] Logout flow

## Phase 22 — Security & Privacy
- [ ] E2E encryption (Signal Protocol atau alternatif)
- [ ] Key exchange protocol
- [ ] Encrypted local storage
- [ ] Input validation & sanitization
- [ ] Rate limiting per endpoint
- [ ] Privacy controls (last seen, read receipts toggle)

## Phase 23 — Performance Optimization
- [ ] FlatList virtualization untuk chat/list panjang
- [ ] Image caching (FastImage / expo-image)
- [ ] Lazy loading untuk tab/screen
- [ ] WebSocket reconnection optimization
- [ ] Memory management (large chat histories)
- [ ] Bundle size optimization

## Phase 24 — Comprehensive Testing
- [ ] Go unit test coverage > 80%
- [ ] Go integration tests (database, API, WebSocket)
- [ ] React Native component tests > 75%
- [ ] React Native hook/store tests
- [ ] E2E tests (Detox/Maestro): auth, chat, document flows
- [ ] Cross-platform tested (iOS + Android)

## Phase 25 — CI/CD Pipeline
- [ ] GitHub Actions: Go lint + test
- [ ] GitHub Actions: RN lint + test
- [ ] Automated builds: iOS (Xcode Cloud / Fastlane) + Android (Gradle)
- [ ] Code quality gates (coverage threshold, lint clean)
- [ ] Staging deployment (server)
- [ ] App distribution (TestFlight + Play Console internal)

## Phase 26 — Beta Release
- [ ] Beta build iOS (TestFlight)
- [ ] Beta build Android (Play Console open testing)
- [ ] Beta server deployed (staging → production-like)
- [ ] Feedback collection system
- [ ] P0/P1 bugs fixed
- [ ] Beta sign-off

## Phase 27 — Production Release
- [ ] Final QA passed (iOS + Android)
- [ ] Production server deployed + monitored
- [ ] App Store submission (review guidelines compliance)
- [ ] Play Store submission
- [ ] Release documentation (CHANGELOG, README)
- [ ] Monitoring & alerting setup (Sentry, uptime)
- [ ] v1.0.0 RELEASED

---

## Status Legend

| Symbol | Meaning |
|--------|---------|
| `[ ]`  | Belum dimulai |
| `[~]`  | Sedang dikerjakan |
| `[x]`  | Selesai |
| `[!]`  | Blocked / ada masalah |

---

## Ringkasan Progress

| Phase | Status | Tanggal Mulai | Tanggal Selesai |
|-------|--------|---------------|-----------------|
| 01    | `[x]`  | 2025-07-11    | 2025-07-11      |
| 02    | `[ ]`  |               |                 |
| 03    | `[ ]`  |               |                 |
| 04    | `[ ]`  |               |                 |
| 05    | `[ ]`  |               |                 |
| 06    | `[ ]`  |               |                 |
| 07    | `[ ]`  |               |                 |
| 08    | `[ ]`  |               |                 |
| 09    | `[ ]`  |               |                 |
| 10    | `[ ]`  |               |                 |
| 11    | `[ ]`  |               |                 |
| 12    | `[ ]`  |               |                 |
| 13    | `[ ]`  |               |                 |
| 14    | `[ ]`  |               |                 |
| 15    | `[ ]`  |               |                 |
| 16    | `[ ]`  |               |                 |
| 17    | `[ ]`  |               |                 |
| 18    | `[ ]`  |               |                 |
| 19    | `[ ]`  |               |                 |
| 20    | `[ ]`  |               |                 |
| 21    | `[ ]`  |               |                 |
| 22    | `[ ]`  |               |                 |
| 23    | `[ ]`  |               |                 |
| 24    | `[ ]`  |               |                 |
| 25    | `[ ]`  |               |                 |
| 26    | `[ ]`  |               |                 |
| 27    | `[ ]`  |               |                 |
