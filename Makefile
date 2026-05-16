.PHONY: help up down build run seed logs test swagger deps tidy security fmt lint db-migrate db-seed

# ── Help ──────────────────────────────────────────────────────
help:
	@echo ""
	@echo "Healthcare API — Make targets"
	@echo "═══════════════════════════════════════════════"
	@echo ""
	@echo "Docker:"
	@echo "  make up       Start all services (docker compose)"
	@echo "  make down     Stop all services"
	@echo "  make logs     Tail logs for all services"
	@echo ""
	@echo "Build / Run:"
	@echo "  make build    Compile Go binary"
	@echo "  make run      Build + run locally"
	@echo "  make deps     Download + tidy dependencies"
	@echo "  make swagger  Regenerate swagger docs (requires swag CLI)"
	@echo ""
	@echo "Database:"
	@echo "  make db-migrate  Apply SQL migrations"
	@echo "  make db-seed     Run development seed (APP_ENV=development)"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  make test     Run all tests"
	@echo "  make fmt      Format code"
	@echo "  make lint     Run golangci-lint"
	@echo "  make security Run security scanners"
	@echo ""

# ── Docker ────────────────────────────────────────────────────
up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

up-dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# ── Build ─────────────────────────────────────────────────────
build:
	@echo "Building healthcare-api..."
	go build -o healthcare-api ./cmd

run: build
	./healthcare-api

# ── Dependencies ─────────────────────────────────────────────
deps:
	@echo "Fetching new dependencies..."
	go get github.com/redis/go-redis/v9@latest
	go get github.com/go-redis/redis_rate/v10@latest
	go get github.com/swaggo/swag@latest
	go get github.com/swaggo/http-swagger@latest
	go mod tidy
	go mod download
	go mod verify

tidy:
	go mod tidy

# ── Swagger ───────────────────────────────────────────────────
swagger:
	@which swag > /dev/null || go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/main.go -o docs --parseDependency

# ── Database ──────────────────────────────────────────────────
db-migrate:
	@echo "Applying migrations..."
	psql -h ${DB_HOST:-localhost} -U ${DB_USER:-postgres} -d ${DB_NAME:-healthcare} \
	     -f migrations/001_create_tables.sql

db-seed:
	@echo "Running bulk seed (development only)..."
	APP_ENV=development go run ./seed/...

# ── Testing ───────────────────────────────────────────────────
test:
	go test ./...

test-v:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-race:
	go test -race ./...

# ── Quality ───────────────────────────────────────────────────
fmt:
	go fmt ./...

lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

security:
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

vet:
	go vet ./...

clean:
	rm -f healthcare-api healthcare-api.exe coverage.out coverage.html
	go clean
