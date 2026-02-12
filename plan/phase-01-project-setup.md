# Phase 01: Project Setup

> Inisialisasi Go backend server dan React Native mobile app.
> Phase ini menghasilkan skeleton project yang bisa di-run di kedua sisi.

**Estimasi:** 3 hari
**Dependency:** Tidak ada
**Output:** Go server berjalan di localhost:8080, React Native app berjalan di simulator/emulator.

---

## Task 1.1: Initialize Go Backend

**Input:** Tidak ada (fresh project)
**Output:** Go project dengan module, folder structure, dan basic HTTP server

### Steps:
1. Buat folder `server/` di root project
2. Run `go mod init github.com/otoritech/chatat` di dalam `server/`
3. Install dependencies awal:
   ```bash
   go get github.com/go-chi/chi/v5
   go get github.com/go-chi/cors
   go get github.com/jackc/pgx/v5
   go get github.com/redis/go-redis/v9
   go get github.com/golang-jwt/jwt/v5
   go get github.com/google/uuid
   go get github.com/rs/zerolog
   go get github.com/gorilla/websocket
   go get golang.org/x/crypto
   go get github.com/joho/godotenv
   ```
4. Buat folder structure sesuai `docs/project-structure.md`:
   ```
   server/
   ├── cmd/
   │   └── server/
   │       └── main.go
   ├── internal/
   │   ├── config/
   │   │   └── config.go
   │   ├── handler/
   │   ├── middleware/
   │   ├── model/
   │   ├── repository/
   │   ├── service/
   │   ├── ws/
   │   └── errors/
   │       └── errors.go
   ├── pkg/
   │   └── response/
   │       └── response.go
   ├── migrations/
   ├── go.mod
   └── go.sum
   ```
5. Buat `cmd/server/main.go` dengan:
   - Load `.env` via godotenv
   - Initialize config dari environment variables
   - Setup zerolog logger
   - Create basic Chi router dengan health check endpoint
   - Start HTTP server di port dari config (default 8080)
6. Buat `internal/config/config.go`:
   ```go
   type Config struct {
       Port        string
       DatabaseURL string
       RedisURL    string
       JWTSecret   string
       Environment string // development, staging, production
   }
   ```
7. Buat `internal/errors/errors.go` dengan `AppError` struct dasar
8. Buat `pkg/response/response.go` dengan JSON response helpers:
   - `Success(w, data, statusCode)`
   - `Error(w, err, statusCode)`
9. Test: `go run cmd/server/main.go` — server berjalan, `GET /health` return 200

### Acceptance Criteria:
- [x] `go build ./...` sukses tanpa error
- [x] `go vet ./...` clean
- [x] Server berjalan di port 8080
- [x] `GET /health` return `{"status": "ok"}`
- [x] Config loaded dari `.env`
- [x] Zerolog output terlihat di console

---

## Task 1.2: Initialize React Native App

**Input:** Tidak ada (fresh project)
**Output:** React Native app yang berjalan di simulator/emulator

### Steps:
1. Buat React Native project di folder `mobile/`:
   ```bash
   npx @react-native-community/cli init Chatat --directory mobile
   ```
   Atau jika menggunakan Expo:
   ```bash
   npx create-expo-app mobile --template blank-typescript
   ```
2. Verify app berjalan:
   - iOS: `npx react-native run-ios` atau `npx expo run:ios`
   - Android: `npx react-native run-android` atau `npx expo run:android`
3. Install core dependencies:
   ```bash
   npm install @react-navigation/native @react-navigation/bottom-tabs @react-navigation/native-stack
   npm install react-native-screens react-native-safe-area-context
   npm install zustand
   npm install react-native-mmkv
   npm install react-native-reanimated react-native-gesture-handler
   npm install @react-native-async-storage/async-storage
   npm install axios
   npm install react-i18next i18next
   npm install date-fns
   ```
4. Buat folder structure sesuai `docs/project-structure.md`:
   ```
   mobile/
   ├── src/
   │   ├── components/
   │   │   ├── ui/
   │   │   ├── layout/
   │   │   └── shared/
   │   ├── screens/
   │   │   ├── auth/
   │   │   ├── chat/
   │   │   ├── document/
   │   │   ├── topic/
   │   │   ├── contact/
   │   │   └── settings/
   │   ├── hooks/
   │   ├── stores/
   │   ├── services/
   │   │   ├── api/
   │   │   └── ws/
   │   ├── types/
   │   ├── utils/
   │   ├── i18n/
   │   ├── theme/
   │   └── navigation/
   ├── App.tsx
   └── package.json
   ```
5. Setup TypeScript strict mode di `tsconfig.json`:
   ```json
   {
     "compilerOptions": {
       "strict": true,
       "noUncheckedIndexedAccess": true,
       "noImplicitReturns": true,
       "forceConsistentCasingInFileNames": true
     }
   }
   ```
6. Buat placeholder `App.tsx` yang menampilkan "Chatat" di layar
7. Verify hot reload berjalan

### Acceptance Criteria:
- [x] App berjalan di iOS simulator
- [x] App berjalan di Android emulator
- [x] TypeScript strict mode aktif tanpa error
- [x] Hot reload berjalan
- [x] Folder structure sesuai panduan

---

## Task 1.3: Docker Development Environment

**Input:** Task 1.1 selesai
**Output:** Docker Compose dengan PostgreSQL dan Redis

### Steps:
1. Buat `docker-compose.yml` di root:
   ```yaml
   version: '3.8'
   services:
     postgres:
       image: postgres:16-alpine
       environment:
         POSTGRES_DB: chatat
         POSTGRES_USER: chatat
         POSTGRES_PASSWORD: chatat_dev
       ports:
         - "5432:5432"
       volumes:
         - pgdata:/var/lib/postgresql/data

     redis:
       image: redis:7-alpine
       ports:
         - "6379:6379"
       volumes:
         - redisdata:/data

   volumes:
     pgdata:
     redisdata:
   ```
2. Buat `.env` file untuk server:
   ```env
   PORT=8080
   DATABASE_URL=postgres://chatat:chatat_dev@localhost:5432/chatat?sslmode=disable
   REDIS_URL=redis://localhost:6379
   JWT_SECRET=dev-secret-change-in-production
   ENVIRONMENT=development
   ```
3. `docker-compose up -d` — PostgreSQL dan Redis berjalan
4. Verify koneksi dari Go server ke PostgreSQL dan Redis
5. Buat `Makefile` dengan commands:
   ```makefile
   .PHONY: dev db-up db-down test lint

   dev:
   	go run cmd/server/main.go

   db-up:
   	docker-compose up -d

   db-down:
   	docker-compose down

   test:
   	go test ./... -v -count=1

   lint:
   	golangci-lint run ./...
   ```

### Acceptance Criteria:
- [x] `docker-compose up -d` berjalan tanpa error
- [x] PostgreSQL accessible di port 5433 (host) → 5432 (container)
- [x] Redis accessible di port 6380 (host) → 6379 (container)
- [x] Go server bisa connect ke PostgreSQL
- [x] Go server bisa connect ke Redis
- [x] `make dev` menjalankan server
- [x] `.env` file ada dan terbaca

---

## Task 1.4: Configure Development Tools

**Input:** Task 1.1, 1.2 selesai
**Output:** Linter, formatter, dan tooling lengkap

### Steps:
1. Setup `golangci-lint` untuk Go:
   - Install: `brew install golangci-lint`
   - Buat `.golangci.yml`:
     ```yaml
     run:
       timeout: 5m
     linters:
       enable:
         - errcheck
         - gosimple
         - govet
         - ineffassign
         - staticcheck
         - unused
         - gofmt
         - goimports
     ```
2. Setup ESLint + Prettier untuk React Native:
   - Install:
     ```bash
     npm install -D eslint @typescript-eslint/eslint-plugin @typescript-eslint/parser
     npm install -D prettier eslint-config-prettier
     ```
   - Buat `.eslintrc.js` dan `.prettierrc`
3. Buat `.editorconfig`:
   ```
   root = true
   [*]
   indent_style = space
   indent_size = 2
   end_of_line = lf
   charset = utf-8
   trim_trailing_whitespace = true
   insert_final_newline = true
   [*.go]
   indent_style = tab
   indent_size = 4
   ```
4. Setup `.gitignore` sesuai `docs/git-workflow.md`:
   - Go binaries, vendor/
   - node_modules/, .expo/
   - .env files
   - Build artifacts
   - IDE files
5. Buat root `README.md` dengan:
   - Project overview
   - Prerequisites (Go 1.23+, Node.js 20+, Docker)
   - Setup instructions
   - Development commands
6. Commit initial project: `chore: initial project setup`

### Acceptance Criteria:
- [x] `golangci-lint run ./...` clean (v2.9.0)
- [x] `npx eslint src/` clean (flat config)
- [x] `.editorconfig` aktif
- [x] `.gitignore` mencakup semua artifacts
- [x] README.md berisi setup instructions
- [x] Initial commit terbuat (3 split commits)

---

## Phase 01 Review

### Testing Checklist:
- [x] Go server — `make dev` berjalan, health check OK
- [x] React Native — app muncul di simulator/emulator
- [x] Docker — PostgreSQL + Redis berjalan
- [x] Database connection — Go bisa query PostgreSQL
- [x] Redis connection — Go bisa ping Redis
- [x] Hot reload — edit Go/RN, lihat perubahan
- [x] Lint — semua linter clean

### Review Checklist:
- [x] Folder structure sesuai `docs/project-structure.md`
- [x] Dependencies sesuai spesifikasi
- [x] Naming sesuai `docs/naming-conventions.md`
- [x] Git commit message sesuai `docs/git-workflow.md`
- [x] Tidak ada TODO atau placeholder yang menggantung
- [x] `.env.example` ada sebagai template
