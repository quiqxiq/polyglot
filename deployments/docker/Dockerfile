# syntax=docker/dockerfile:1
# ─── Multi-Stage Dockerfile for Polyglot NetOps Engine ───

# 1. Base Stage: Shared Go Environment
FROM golang:alpine AS base
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download

# 2. Development Stage (Live reload / Dev execution)
FROM base AS dev
COPY . .
EXPOSE 8080
CMD ["go", "run", "./cmd/server"]

# 3. Builder Stage: Compile static production binaries
FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/server ./cmd/server && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /out/seed ./scripts/seed.go

# 4. Production Stage: Minimal, hardened Alpine container
FROM alpine:3.20 AS prod
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app
COPY --from=builder /out/server /usr/local/bin/server
COPY --from=builder /out/seed /usr/local/bin/seed

# Copy internal configuration/templates needed at runtime
COPY --from=builder /app/internal/templates /app/internal/templates

USER appuser
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/server"]
