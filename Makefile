.PHONY: dev db-up db-down db-reset test lint fmt build clean migrate-up migrate-down migrate-create test-go test-mobile test-coverage test-report docker-build

# --- Server ---

dev:
	cd server && go run cmd/server/main.go

build:
	cd server && go build -o bin/chatat cmd/server/main.go

clean:
	rm -rf server/bin

test: test-go test-mobile

test-go:
	cd server && go test ./... -short -count=1 -race -cover

test-mobile:
	cd mobile && npx jest --no-coverage

test-coverage:
	cd server && go test ./... -short -count=1 -race -coverprofile=coverage.out
	cd mobile && npx jest --coverage

test-report:
	bash scripts/test-report.sh

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

docker-build:
	docker build -t chatat-server:latest server/

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

check: vet lint test mobile-typecheck mobile-lint
