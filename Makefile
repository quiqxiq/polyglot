.PHONY: build test test-integration lint fmt run \
        db-up db-down db-logs db-reset \
        migrate-install migrate-up migrate-down migrate-force migrate-create

# ─── Ensure go install'd tools (migrate, golangci-lint) are on PATH ───
# go install places binaries in $(go env GOPATH)/bin, which is not always
# on the user's shell PATH. Prepend it here so every target finds them.
PATH := $(shell go env GOPATH)/bin:$(PATH)

# ─── Load .env if present ─────────────────────────────────────────────
# Makes integration-test env vars (MIKROTIK_TEST_*, DATABASE_URL, etc.)
# visible to every target. .env is gitignored; .env.example is the tracked
# template. Directives must start at column 0 (no leading whitespace).
ifneq (,$(wildcard .env))
include .env
export
endif

# ─── Defaults (overridable via env or .env) ───────────────────────────
# Match deployments/docker-compose.yml so `make db-up && make migrate-up`
# works with zero configuration.
POSTGRES_HOST     ?= localhost
POSTGRES_PORT     ?= 5432
POSTGRES_USER     ?= postgres
POSTGRES_PASSWORD ?= netops
POSTGRES_DB       ?= netops
DATABASE_URL      ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATIONS_PATH   ?= migrations
MIGRATE_VERSION   ?= v4.19.1
COMPOSE_FILE      ?= deployments/docker-compose.yml

# ─── Go ───────────────────────────────────────────────────────────────
build:
	go build ./...

test:
	go test ./... -race -cover

# Butuh perangkat/servis eksternal asli (Mikrotik CHR, GenieACS) yang bisa
# dijangkau — lihat test/integration/*_test.go untuk env var yang wajib
# di-set. Tidak dijalankan oleh `make test` biasa (build tag "integration").
test-integration:
	go test -tags=integration ./test/integration/... -v

lint:
	golangci-lint run ./...

fmt:
	gofmt -l .
	gofumpt -l . || true

run:
	go run ./cmd/server

# ─── Database (Docker Compose) ────────────────────────────────────────
# db-up and db-reset wait for Postgres to accept connections before
# returning, so a subsequent `make migrate-up` never races a container
# that is still initializing its data directory.
db-up:
	docker compose -f $(COMPOSE_FILE) up -d postgres
	@echo "Waiting for Postgres..."
	@for i in $$(seq 1 30); do \
		docker compose -f $(COMPOSE_FILE) exec -T postgres pg_isready -U postgres >/dev/null 2>&1 && echo "ready" && break; \
		sleep 1; \
	done

db-down:
	docker compose -f $(COMPOSE_FILE) down

db-logs:
	docker compose -f $(COMPOSE_FILE) logs -f postgres

# Reset: tear down, remove the volume, bring back up. Destroys data.
db-reset:
	docker compose -f $(COMPOSE_FILE) down -v
	docker compose -f $(COMPOSE_FILE) up -d postgres
	@echo "Waiting for Postgres..."
	@for i in $$(seq 1 30); do \
		docker compose -f $(COMPOSE_FILE) exec -T postgres pg_isready -U postgres >/dev/null 2>&1 && echo "ready" && break; \
		sleep 1; \
	done

# ─── Migrations (golang-migrate) ──────────────────────────────────────
# Install the CLI once:  make migrate-install
# Then:                   make migrate-up
#
# migrate is expected on PATH after `make migrate-install`. If you see
# "migrate: not found", run that target first.
migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)

# Apply all pending migrations.
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" up

# Roll back exactly one migration (safest default — use `migrate down -all`
# directly if you genuinely want to wipe everything).
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" down 1

# Force-set the migration version (use after a botched manual edit):
#   make migrate-force V=3
migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DATABASE_URL)" force $(V)

# Create a new migration pair (up + down). Example:
#   make migrate-create NAME=create_customers_table
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(NAME)
