# Chatat

> WhatsApp-style chat + Notion-style document collaboration mobile app.

## Tech Stack

- **Backend:** Go 1.23+ (Chi, pgx, Redis, WebSocket, JWT, zerolog)
- **Frontend:** React Native / Expo (TypeScript, Zustand, React Navigation)
- **Database:** PostgreSQL 16, Redis 7
- **Infrastructure:** Docker Compose

## Prerequisites

- Go 1.23+
- Node.js 20+
- Docker & Docker Compose
- Xcode (for iOS)
- Android Studio (for Android)

## Getting Started

### 1. Clone & Setup

```bash
git clone https://github.com/otoritech/chatat.git
cd chatat
```

### 2. Start Database Services

```bash
make db-up
```

This starts PostgreSQL (port 5433) and Redis (port 6380) via Docker Compose.

### 3. Start Backend Server

```bash
cp server/.env.example server/.env
make dev
```

Server runs at `http://localhost:8080`. Health check: `GET /health`.

### 4. Start Mobile App

```bash
cd mobile
npm install
npx expo start
```

### 5. Verify Setup

```bash
# Backend health check
curl http://localhost:8080/health

# Run tests
make test

# Run linter
make lint

# TypeScript check
make mobile-typecheck
```

## Development Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start Go server |
| `make build` | Build Go binary |
| `make test` | Run Go tests |
| `make lint` | Run golangci-lint |
| `make vet` | Run go vet |
| `make fmt` | Format Go code |
| `make db-up` | Start PostgreSQL + Redis |
| `make db-down` | Stop PostgreSQL + Redis |
| `make db-reset` | Reset database (drop + recreate) |
| `make mobile-start` | Start Expo dev server |
| `make mobile-typecheck` | TypeScript type check |
| `make check` | Run all checks (vet + lint + test + typecheck) |

## Project Structure

```
chatat/
├── server/          # Go backend (API + WebSocket)
├── mobile/          # React Native / Expo frontend
├── docs/            # Development documentation
├── plan/            # Implementation plan (27 phases)
└── docker-compose.yml
```

See [docs/project-structure.md](docs/project-structure.md) for detailed structure.

## Documentation

- [Project Structure](docs/project-structure.md)
- [Naming Conventions](docs/naming-conventions.md)
- [Error Handling](docs/error-handling.md)
- [Design Patterns](docs/design-patterns.md)
- [Git Workflow](docs/git-workflow.md)
- [Testing Strategy](docs/testing-strategy.md)
- [Go Style Guide](docs/go-style-guide.md)
- [React Native Style Guide](docs/react-native-style-guide.md)

## License

Private - All rights reserved.
