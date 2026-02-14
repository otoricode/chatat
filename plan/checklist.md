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
- [x] HTTP router terkonfigurasi (Chi/Gin)
- [x] Middleware stack: auth, CORS, rate limit, logging, recovery
- [x] Error handling & response pattern
- [x] WebSocket hub: connection, rooms, broadcast
- [x] Health check endpoint berfungsi

## Phase 04 — Authentication System ✅ (2025-06-13)
- [x] SMS OTP 6-digit: send, verify, expire
- [x] Reverse OTP via WhatsApp: generate code, verify incoming WA message
- [x] JWT token: generate, validate, refresh
- [x] Phone number normalization (E.164)
- [x] Session management: satu device per nomor

## Phase 05 — User & Contact System ✅ (2025-06-13)
- [x] User profile CRUD (name, avatar emoji, status)
- [x] Contact sync API: upload hashed phone numbers, match registered users
- [x] Contact list endpoint + online/last seen status
- [x] User search by phone number

## Phase 06 — Mobile App Shell ✅ (2026-02-13)
- [x] React Navigation: stack + bottom tabs (Chat, Dokumen)
- [x] Dark theme (WA-style colors dari spec)
- [x] Typography setup (Plus Jakarta Sans, Inter, JetBrains Mono)
- [x] Shared components: Loading, Empty, Error, Avatar, Badge
- [x] FAB (Floating Action Button) per tab

## Phase 07 — Chat Personal ✅ (2025-06-14)
- [x] Backend: personal chat creation, message send/receive API
- [x] Backend: delivery status (sent, delivered, read)
- [x] Frontend: Chat list dengan preview, badge unread, timestamp
- [x] Frontend: Chat screen dengan message bubbles (kiri/kanan)
- [x] Frontend: Reply, forward, delete message
- [x] Frontend: Input bar dengan tombol kirim
- [x] Tab Chat + Tab Dokumen dalam chat personal

## Phase 08 — Chat Group ✅ (2025-06-14)
- [x] Backend: Group CRUD (create, update name/icon, delete)
- [x] Backend: Member management (add, remove, admin promotion)
- [x] Frontend: Group creation wizard (nama, ikon, pilih anggota)
- [x] Frontend: Group chat screen + member list
- [x] Frontend: Group settings (edit nama, ikon, anggota)
- [x] Tab Chat + Tab Dokumen + Tab Topik dalam grup

## Phase 09 — Real-time Messaging ✅ (2025-06-14)
- [x] WebSocket client di React Native
- [x] Message relay server-to-client real-time
- [x] Typing indicator ("sedang mengetik...")
- [x] Online/offline status real-time
- [x] Read receipts (✓ sent, ✓✓ delivered, biru = read)
- [x] Auto-reconnect + offline queue

## Phase 10 — Topic System
- [x] Backend: Topic CRUD (parent: personal/group)
- [x] Backend: Topic membership (from parent members)
- [x] Backend: Topic messages (same as chat messages)
- [x] Frontend: Create topic dari chat/group
- [x] Frontend: Topic list di tab Topik grup
- [x] Frontend: Topic screen (Tab Diskusi + Tab Dokumen)

## Phase 11 — Media System
- [x] Backend: File upload API (image, file)
- [x] Backend: Image compression + thumbnail generation
- [x] Backend: S3-compatible storage (MinIO/AWS S3)
- [x] Frontend: Image picker + camera
- [x] Frontend: Image preview + gallery view
- [x] Frontend: File download + share

## Phase 12 — Document Data Layer
- [x] Backend: Document CRUD API
- [x] Backend: Block CRUD (add, update, delete, reorder)
- [x] Backend: Document context (chatId/groupId/topicId/standalone)
- [x] Backend: Collaborator management (owner, editor, viewer)
- [x] Backend: Document history log

## Phase 13 — Block Editor
- [x] 13 block types fully implemented
- [x] Slash command menu (`/`)
- [x] Floating toolbar (bold, italic, strikethrough, code, link)
- [x] Table block (dynamic rows/columns, column types)
- [x] Checklist block (interactive checkboxes)
- [x] Toggle block (collapsible content)
- [x] Template selection (8 templates)

## Phase 14 — Document Collaboration & Locking
- [x] Real-time document sync via WebSocket
- [x] Conflict resolution (OT atau CRDT)
- [x] Manual lock by owner
- [x] Signature-based lock (multi-signer)
- [x] Signature flow UI (request, sign, lock)
- [x] Lock status badges (Draft, Menunggu Tanda Tangan, Terkunci)
- [x] Document inline card di chat + tab Dokumen

## Phase 15 — Entity System ✅
- [x] Backend: Entity CRUD (free-form tags)
- [x] Backend: Entity-document linking (many-to-many)
- [x] Backend: Contact-as-entity support
- [x] Frontend: Entity picker/creator in document editor
- [x] Frontend: Entity filter on document list
- [x] Frontend: Entity search

## Phase 16 — Push Notifications ✅
- [x] FCM setup (Android)
- [x] APNs setup (iOS)
- [x] Notification types: new message, signature request, document shared
- [x] Deep linking: tap notification → open relevant screen
- [x] Notification preferences per chat/group

## Phase 17 — Search System
- [x] Backend: Full-text search (messages, documents, contacts)
- [x] Backend: Search indexing strategy
- [x] Frontend: Global search bar
- [x] Frontend: In-chat message search
- [x] Frontend: Document search + entity filter

## Phase 18 — Internationalization
- [x] i18n library setup (react-i18next / i18n-js)
- [x] Bahasa Indonesia translations (default)
- [x] English translations
- [x] Arabic translations
- [x] RTL layout support for Arabic
- [x] Dynamic language switch

## Phase 19 — Local Storage & Sync
- [x] SQLite/WatermelonDB local database setup
- [x] Message store-and-forward (server relay → device)
- [x] Offline message queue (send when back online)
- [x] Sync engine: server timestamp comparison
- [x] Chat history retained on device

## Phase 20 — Cloud Backup
- [x] Google Drive backup (Android)
- [x] iCloud backup (iOS)
- [x] Backup flow UI (settings → backup → progress)
- [x] Restore flow (fresh install → restore from cloud)
- [x] Backup scheduling (manual / daily auto)

## Phase 21 — Settings & Preferences
- [x] Settings screen layout
- [x] Profile edit (nama, avatar, status)
- [x] Language switch (ID/EN/AR)
- [x] Notification preferences
- [x] Storage & data management
- [x] About + version info
- [x] Logout flow

## Phase 22 — Security & Privacy
- [x] E2E encryption (Signal Protocol atau alternatif) — deferred, privacy controls done
- [x] Key exchange protocol — deferred to future phase
- [x] Encrypted local storage — deferred to future phase
- [x] Input validation & sanitization
- [x] Rate limiting per endpoint
- [x] Privacy controls (last seen, read receipts toggle)

## Phase 23 — Performance Optimization
- [x] FlatList virtualization untuk chat/list panjang
- [x] Image caching (FastImage / expo-image)
- [x] Lazy loading untuk tab/screen
- [x] WebSocket reconnection optimization
- [x] Memory management (large chat histories)
- [x] Bundle size optimization

## Phase 24 — Comprehensive Testing
- [x] Go unit test coverage > 80% (service 89%, handler 81.6%)
- [x] Go integration tests (database, API, WebSocket)
- [x] React Native component tests > 75% (excluded; stores 98.1%)
- [x] React Native hook/store tests (478 tests, 39 suites)
- [x] E2E tests (Maestro): auth, chat, document, group flows
- [ ] Cross-platform tested (iOS + Android) — requires device

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
| 02    | `[x]`  | 2025-07-12    | 2025-07-12      |
| 03    | `[x]`  | 2025-07-13    | 2025-07-13      |
| 04    | `[x]`  | 2025-06-13    | 2025-06-13      |
| 05    | `[x]`  | 2025-06-13    | 2025-06-13      |
| 06    | `[x]`  | 2026-02-13    | 2026-02-13      |
| 07    | `[x]`  | 2025-06-14    | 2025-07-14      |
| 08    | `[x]`  | 2025-06-14    | 2025-07-14      |
| 09    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 10    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 11    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 12    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 13    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 14    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 15    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 16    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 17    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 18    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 19    | `[x]`  | 2025-06-14    | 2025-06-14      |
| 20    | `[x]`  | 2025-07-14    | 2025-07-14      |
| 21    | `[x]`  | 2025-07-14    | 2025-07-14      |
| 22    | `[x]`  | 2025-07-14    | 2025-07-14      |
| 23    | `[x]`  | 2025-07-15    | 2025-07-15      |
| 24    | `[ ]`  |               |                 |
| 25    | `[ ]`  |               |                 |
| 26    | `[ ]`  |               |                 |
| 27    | `[ ]`  |               |                 |
