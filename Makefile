.PHONY: dev db-up db-down db-reset test lint fmt build clean migrate-up migrate-down migrate-create

# --- Server ---

dev:
	cd server && go run cmd/server/main.go

build:
	cd server && go build -o bin/chatat cmd/server/main.go

clean:
	rm -rf server/bin

test:
	cd server && go test ./... -v -count=1

lint:
	cd server && golangci-lint run ./...

fmt:
	cd server && gofmt -w . && goimports -w .

vet:
	cd server && go vet ./...

# --- Docker ---

db-up:
	docker compose up -d

db-down:
	docker compose down

db-reset:
	docker compose down -v && docker compose up -d

# --- Migrations ---

migrate-up:
	migrate -path server/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path server/migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir server/migrations -seq $(name)

# --- Mobile ---

mobile-start:
	cd mobile && npx expo start

mobile-ios:
	cd mobile && npx expo run:ios

mobile-android:
	cd mobile && npx expo run:android

mobile-lint:
	cd mobile && npx eslint src/ --ext .ts,.tsx

mobile-typecheck:
	cd mobile && npx tsc --noEmit

# --- All ---

check: vet lint test mobile-typecheck
