.PHONY: build test test-integration lint fmt run setup seed \
        proto proto-tools proto-clean \
        dev-up dev-down dev-logs dev-setup \
        prod-build prod-up prod-down prod-logs prod-setup \
        db-up db-down db-logs db-reset db-wait \
        migrate-install migrate-up migrate-down migrate-force migrate-create

# ─── Ensure go install'd tools (migrate, golangci-lint, protoc-gen-*) are on PATH ───
PATH := $(shell go env GOPATH)/bin:$(PATH)

# ─── Load .env if present ─────────────────────────────────────────────
ifneq (,$(wildcard .env))
include .env
export
endif

# ─── Defaults (overridable via env or .env) ───────────────────────────
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
POSTGRES_USER     ?= postgres
POSTGRES_PASSWORD ?= netops
POSTGRES_DB       ?= netops
DATABASE_URL      ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_PATH   ?= migrations
MIGRATE_VERSION   ?= v4.19.1
COMPOSE_DEV       ?= deployments/docker-compose.yml
COMPOSE_PROD      ?= deployments/docker-compose.prod.yml
PROTO_ROOT        ?= api/proto
PROTO_OUT         ?= api/gen

# ─── General Shortcuts ────────────────────────────────────────────────
setup: dev-setup

# ─── Go Targets ───────────────────────────────────────────────────────
build:
	go build ./...

test:
	go test ./... -race -cover

test-integration:
	go test -tags=integration ./test/integration/... -v

lint:
	golangci-lint run ./...

fmt:
	gofmt -l .
	gofumpt -l . || true

run:
	go run ./cmd/server

# ─── Protobuf / gRPC & ConnectRPC Web (api/proto/v1) ───────────────────
PROTO_FILES := $(shell find $(PROTO_ROOT) -name '*.proto' 2>/dev/null)
WEB_GEN_OUT := website/src/gen

proto:
	mkdir -p $(PROTO_OUT)
	protoc \
		--proto_path=$(PROTO_ROOT) \
		--go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)

proto-web:
	mkdir -p $(WEB_GEN_OUT)
	protoc \
		--proto_path=$(PROTO_ROOT) \
		--plugin=protoc-gen-es=website/node_modules/.bin/protoc-gen-es \
		--es_out=$(WEB_GEN_OUT) --es_opt=target=ts \
		--plugin=protoc-gen-connect-es=website/node_modules/.bin/protoc-gen-connect-es \
		--connect-es_out=$(WEB_GEN_OUT) --connect-es_opt=target=ts \
		$(PROTO_FILES)

proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

proto-clean:
	rm -rf $(PROTO_OUT) $(WEB_GEN_OUT)

# ─── Seeding ──────────────────────────────────────────────────────────
seed:
	DATABASE_URL="$(DATABASE_URL)" go run ./scripts/seed.go

# ─── Database & Cache Infrastructure (Postgres + Redis) ───────────────
db-up:
	docker compose -f $(COMPOSE_DEV) up -d postgres redis
	@$(MAKE) db-wait

db-wait:
	@echo "Waiting for Postgres and Redis..."
	@for i in $$(seq 1 30); do \
		docker compose -f $(COMPOSE_DEV) exec -T postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) >/dev/null 2>&1 && echo "Postgres is ready!" && break; \
		sleep 1; \
	done
	@for i in $$(seq 1 30); do \
		docker compose -f $(COMPOSE_DEV) exec -T redis redis-cli ping >/dev/null 2>&1 && echo "Redis is ready!" && break; \
		sleep 1; \
	done

db-down:
	docker compose -f $(COMPOSE_DEV) down

db-logs:
	docker compose -f $(COMPOSE_DEV) logs -f postgres redis

db-reset:
	docker compose -f $(COMPOSE_DEV) down -v
	docker compose -f $(COMPOSE_DEV) up -d postgres redis
	@$(MAKE) db-wait

# ─── Migrations (golang-migrate) ──────────────────────────────────────
migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(V)

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)

# ─── Development Stack (Docker Compose Dev) ───────────────────────────
dev-up:
	docker compose -f $(COMPOSE_DEV) up -d

dev-down:
	docker compose -f $(COMPOSE_DEV) down

dev-logs:
	docker compose -f $(COMPOSE_DEV) logs -f

dev-setup:
	@echo "=== [DEV SETUP] Starting Postgres & Redis ==="
	docker compose -f $(COMPOSE_DEV) up -d postgres redis
	@$(MAKE) db-wait
	@echo "=== [DEV SETUP] Running Migrations ==="
	@$(MAKE) migrate-up
	@echo "=== [DEV SETUP] Seeding Database ==="
	@$(MAKE) seed
	@echo "=== [DEV SETUP] Starting Dev Server & Web Containers ==="
	docker compose -f $(COMPOSE_DEV) up -d server web
	@echo "=== Dev Environment Setup Completed! ==="

# ─── Production Stack (Docker Compose Prod) ───────────────────────────
prod-build:
	docker build --target prod -t polyglot:latest -f deployments/docker/Dockerfile .

prod-up:
	docker compose -f $(COMPOSE_PROD) up -d

prod-down:
	docker compose -f $(COMPOSE_PROD) down

prod-logs:
	docker compose -f $(COMPOSE_PROD) logs -f

prod-setup:
	@echo "=== [PROD SETUP] Starting Production Postgres & Redis ==="
	docker compose -f $(COMPOSE_PROD) up -d postgres redis
	@echo "Waiting for Production Postgres..."
	@for i in $$(seq 1 30); do \
		docker compose -f $(COMPOSE_PROD) exec -T postgres pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) >/dev/null 2>&1 && echo "Postgres is ready" && break; \
		sleep 1; \
	done
	@echo "=== [PROD SETUP] Running Migrations ==="
	@$(MAKE) migrate-up
	@echo "=== [PROD SETUP] Seeding Database ==="
	@$(MAKE) seed
	@echo "=== [PROD SETUP] Starting Production Server ==="
	docker compose -f $(COMPOSE_PROD) up -d server
	@echo "=== Production Stack Setup Completed ==="