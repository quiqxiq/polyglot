# syntax=docker/dockerfile:1
#
# CATATAN: cek tag `golang:1.26-alpine` benar-benar tersedia di
# Docker Hub saat build (patch terbaru per TECH-STACK-DAN-PERSIAPAN.md §1
# adalah 1.26.4) — kalau tag ini belum ada, ganti ke tag patch spesifik
# mis. `golang:1.26.4-alpine`.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
COPY --from=build /out/server /usr/local/bin/server
ENTRYPOINT ["/usr/local/bin/server"]
