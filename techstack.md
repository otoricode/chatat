# Tech Stack — Chatat

> Seluruh teknologi, library, dan tools yang digunakan dalam project Chatat.
> Berdasarkan plan Phase 01–27.

---

## Backend

### Runtime & Language

| Technology | Version | Purpose | Phase |
|---|---|---|---|
| Go | 1.23+ | Bahasa utama backend | 01 |
| Air | latest | Hot reload development | 01 |

### Framework & Router

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Chi | `github.com/go-chi/chi/v5` | HTTP router + middleware | 03 |
| Chi CORS | `github.com/go-chi/cors` | CORS middleware | 03, 22 |
| Chi Compress | built-in middleware | Response gzip compression | 23 |

### Database

| Technology | Version | Purpose | Phase |
|---|---|---|---|
| PostgreSQL | 16+ | Primary relational database | 02 |
| pgx | `github.com/jackc/pgx/v5` | PostgreSQL driver (native, non-ORM) | 02 |
| pgxpool | `github.com/jackc/pgx/v5/pgxpool` | Connection pooling | 02, 23 |
| golang-migrate | `github.com/golang-migrate/migrate/v4` | Database migration tool | 02 |
| uuid-ossp | PostgreSQL extension | UUID generation | 02 |
| tsvector/tsquery | PostgreSQL built-in | Full-text search (Indonesian) | 17 |

### Cache & In-Memory

| Technology | Version | Purpose | Phase |
|---|---|---|---|
| Redis | 7+ | OTP storage, sessions, rate limiting, caching, online status, typing indicators | 02, 04, 09, 22, 23 |
| go-redis | `github.com/redis/go-redis/v9` | Redis client for Go | 02 |

### Authentication & Security

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| golang-jwt | `github.com/golang-jwt/jwt/v5` | JWT access/refresh tokens | 04 |
| bcrypt | `golang.org/x/crypto/bcrypt` | Password/PIN hashing | 04, 14 |
| bluemonday | `github.com/microcosm-cc/bluemonday` | HTML sanitization | 22 |
| validator | `github.com/go-playground/validator/v10` | Input struct validation | 22 |

### Real-time

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| gorilla/websocket | `github.com/gorilla/websocket` | WebSocket server (chat real-time, doc collab) | 03, 09, 14 |

### Storage

| Technology | Purpose | Phase |
|---|---|---|
| MinIO | S3-compatible storage (development) | 11 |
| AWS S3 | Media storage (production) | 11 |
| aws-sdk-go-v2 | `github.com/aws/aws-sdk-go-v2` — S3 client | 11 |

### Image Processing

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| imaging | `github.com/disintegration/imaging` | Image resize, thumbnails, compression | 11 |

### Logging

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| zerolog | `github.com/rs/zerolog` | Structured JSON logging | 03 |

### Push Notifications

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Firebase Admin Go | `firebase.google.com/go/v4` | FCM push notification dispatch | 16 |

### SMS & WhatsApp

| Technology | Purpose | Phase |
|---|---|---|
| SMS Provider API | OTP via SMS (Twilio/Vonage/local) | 04 |
| WhatsApp Business API | Reverse OTP authentication | 04 |

### Testing (Backend)

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| testify | `github.com/stretchr/testify` | Assertions + mocking | 24 |
| mockery | `github.com/vektra/mockery/v2` | Interface mock generation | 24 |
| httptest | `net/http/httptest` (stdlib) | HTTP handler testing | 24 |
| testcontainers-go | `github.com/testcontainers/testcontainers-go` | Docker containers for integration tests | 24 |

### Deployment

| Technology | Purpose | Phase |
|---|---|---|
| Docker | Container build + deployment | 01, 25, 27 |
| Docker Compose | Local development (PostgreSQL, Redis, MinIO) | 01 |
| GitHub Actions | CI/CD automation | 25 |

---

## Mobile (React Native)

### Runtime & Language

| Technology | Version | Purpose | Phase |
|---|---|---|---|
| React Native | 0.75+ | Cross-platform mobile framework | 06 |
| TypeScript | 5.x (strict) | Type safety | 06 |
| Hermes | built-in | JavaScript engine (Android, optimized) | 23 |

### Navigation

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| React Navigation | `@react-navigation/native` | Screen navigation | 06 |
| Stack Navigator | `@react-navigation/stack` | Stack-based navigation | 06 |
| Bottom Tabs | `@react-navigation/bottom-tabs` | Tab bar (Chat + Dokumen) | 06 |
| Material Top Tabs | `@react-navigation/material-top-tabs` | Swipeable tabs in chat (Obrolan, Topik, Dokumen) | 08 |

### State Management

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Zustand | `zustand` | Global state management | 06 |
| zustand/shallow | built-in | Shallow comparison for selectors | 23 |
| React Query | `@tanstack/react-query` | Server state / async data fetching | 07 |

### Networking

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Axios | `axios` | HTTP client for REST API | 06 |
| WebSocket (native) | built-in | Real-time messaging | 09 |
| NetInfo | `@react-native-community/netinfo` | Network status detection | 19 |

### Local Storage

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| MMKV | `react-native-mmkv` | Fast key-value storage (settings, tokens) | 06 |
| AsyncStorage | `@react-native-async-storage/async-storage` | Language preference, settings | 18 |
| SecureStore | `expo-secure-store` | Secure token storage | 22 |
| WatermelonDB | `@nozbe/watermelondb` | Local SQLite database (offline support) | 19 |

### UI Components

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| React Native Reanimated | `react-native-reanimated` | Smooth animations, drag-to-reorder | 13 |
| React Native Gesture Handler | `react-native-gesture-handler` | Touch gestures | 06 |
| React Native Safe Area | `react-native-safe-area-context` | Safe area insets | 06 |
| React Native SVG | `react-native-svg` | SVG icon rendering | 06 |
| FastImage | `react-native-fast-image` | Image caching, priority loading | 23 |
| React Native Image Picker | `react-native-image-picker` | Photo/file picker | 11, 21 |

### Internationalization

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| i18next | `i18next` | i18n framework | 18 |
| react-i18next | `react-i18next` | React bindings for i18n | 18 |
| react-native-localize | `react-native-localize` | Device locale detection | 18 |
| RNRestart | `react-native-restart` | App restart for RTL switch | 18 |

### Push Notifications

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Firebase Messaging | `@react-native-firebase/messaging` | FCM push notification handling | 16 |
| Firebase Crashlytics | `@react-native-firebase/crashlytics` | Crash reporting | 26 |

### Cloud Backup

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Google Sign-In | `@react-native-google-signin/google-signin` | Google auth for Drive access | 20 |
| Google Drive API | `@robinbobin/react-native-google-drive-api-wrapper` | Google Drive backup (Android) | 20 |
| Cloud Store | `react-native-cloud-store` | iCloud storage (iOS) | 20 |

### Testing (Mobile)

| Technology | Package | Purpose | Phase |
|---|---|---|---|
| Jest | `jest` | Test runner | 24 |
| React Native Testing Library | `@testing-library/react-native` | Component testing | 24 |
| jest-native | `@testing-library/jest-native` | Native matchers | 24 |
| Maestro | CLI tool | End-to-end testing | 24 |

### Build & Distribution

| Technology | Purpose | Phase |
|---|---|---|
| Fastlane | Build automation + store upload | 26 |
| Xcode | iOS build toolchain | 26 |
| Gradle | Android build toolchain | 26 |
| CocoaPods | iOS dependency manager | 06 |

---

## DevOps & Infrastructure

### Development

| Technology | Purpose | Phase |
|---|---|---|
| Docker Compose | Local services (PostgreSQL, Redis, MinIO) | 01 |
| Air | Go hot reload | 01 |
| Metro Bundler | React Native bundler | 06 |

### CI/CD

| Technology | Purpose | Phase |
|---|---|---|
| GitHub Actions | Automated testing, building, deployment | 25 |
| golangci-lint | Go linter (CI) | 25 |
| ESLint | TypeScript/React linter (CI) | 25 |
| Codecov | Coverage reporting | 25 |
| commitlint | Conventional commit enforcement | 25 |

### Production

| Technology | Purpose | Phase |
|---|---|---|
| PostgreSQL (managed) | Production database with backups | 27 |
| Redis (managed) | Production cache with persistence | 27 |
| AWS S3 | Production media storage | 27 |
| Docker | Container deployment | 27 |
| SSL/TLS | HTTPS encryption | 27 |
| Sentry (or equivalent) | Error tracking + monitoring | 27 |

### Distribution

| Technology | Purpose | Phase |
|---|---|---|
| Apple TestFlight | iOS beta distribution | 26 |
| Google Play Console | Android beta distribution (internal track) | 26 |
| App Store | iOS production distribution | 27 |
| Google Play Store | Android production distribution | 27 |

---

## Ringkasan Dependensi Utama

### Go Modules (go.mod)

```
github.com/go-chi/chi/v5
github.com/go-chi/cors
github.com/jackc/pgx/v5
github.com/golang-migrate/migrate/v4
github.com/redis/go-redis/v9
github.com/golang-jwt/jwt/v5
github.com/gorilla/websocket
github.com/rs/zerolog
github.com/disintegration/imaging
github.com/go-playground/validator/v10
github.com/microcosm-cc/bluemonday
github.com/aws/aws-sdk-go-v2
firebase.google.com/go/v4
github.com/stretchr/testify
github.com/vektra/mockery/v2
github.com/google/uuid
golang.org/x/crypto
```

### Node Packages (package.json)

```
react-native
typescript
zustand
@tanstack/react-query
axios
react-native-mmkv
@nozbe/watermelondb
@react-navigation/native
@react-navigation/stack
@react-navigation/bottom-tabs
@react-navigation/material-top-tabs
react-native-reanimated
react-native-gesture-handler
react-native-safe-area-context
react-native-svg
react-native-fast-image
react-native-image-picker
i18next
react-i18next
react-native-localize
react-native-restart
@react-native-firebase/messaging
@react-native-firebase/crashlytics
@react-native-google-signin/google-signin
@react-native-community/netinfo
@react-native-async-storage/async-storage
expo-secure-store
react-native-cloud-store
@testing-library/react-native
jest
```

---

## Versi Minimum

| Component | Minimum Version |
|---|---|
| Go | 1.23 |
| Node.js | 20 LTS |
| React Native | 0.75 |
| PostgreSQL | 16 |
| Redis | 7 |
| Xcode | 15+ |
| Android SDK (compileSdk) | 34 |
| Java (Android build) | 17 |
| CocoaPods | 1.14+ |
| iOS deployment target | 14.0 |
| Android minSdk | 24 (Android 7.0) |
