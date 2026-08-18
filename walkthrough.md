# Walkthrough: PPPoE / PPP Management Module

Implementasi modul manajemen **PPPoE / PPP** lengkap telah selesai dibangun dari sisi backend hingga frontend web sesuai arsitektur Clean Architecture dan standar TanStack Table.

---

## 🚀 Ringkasan Fitur yang Dibangun

### 1. Protobuf Contract & Code Generation
- File: [api/proto/v1/ppp.proto](file:///home/quixiq/projects/work/polyglot/api/proto/v1/ppp.proto)
- Mendefinisikan entity:
  - `PPPSecret`: Subscriber credentials, IP pool, service, profile, comment, status.
  - `PPPProfile`: Bandwidth rate-limits (`rx/tx`), local/remote IP addresses, DNS servers, address list, parent queue, and concurrency (`only-one`/`shared-users`).
  - `PPPActiveSession`: Real-time session monitoring with caller-ID/MAC, assigned IP, uptime, traffic bytes, and RADIUS auth flag.
- Mendefinisikan service `PPPService` dengan 17 ConnectRPC / gRPC RPCs (CRUD Secrets, CRUD Profiles, Active/Inactive Monitoring, Bulk Disconnect, and Server Streaming).
- Berhasil di-generate ke Go (`api/gen/v1/ppp*.pb.go`) dan TypeScript (`web/src/gen/v1/ppp*.ts`).

---

### 2. Backend Go Clean Architecture (1.22+ ServeMux)
- **Port Layer**:
  - [internal/port/ppp_session.go](file:///home/quixiq/projects/work/polyglot/internal/port/ppp_session.go): Definisi domain structs (`PPPProfile`, `PPPProfileParams`, `PPPoESecretParams`, `PPPActiveSession`).
  - [internal/port/ppp_gateway.go](file:///home/quixiq/projects/work/polyglot/internal/port/ppp_gateway.go): Kontrak interface `PPPGateway`.
- **Driver Layer**:
  - [internal/driver/mikrotik/gateway.go](file:///home/quixiq/projects/work/polyglot/internal/driver/mikrotik/gateway.go): Implementasi `PPPGateway` mengeksekusi RouterOS commands melalui `CommandExecutor` ber-policy gate.
  - [internal/driver/mikrotik/ppp.go](file:///home/quixiq/projects/work/polyglot/internal/driver/mikrotik/ppp.go) & [ppp_profile.go](file:///home/quixiq/projects/work/polyglot/internal/driver/mikrotik/ppp_profile.go): Command builders dan parsers.
- **Usecase Layer**:
  - [internal/usecase/ppp/manage_ppp.go](file:///home/quixiq/projects/work/polyglot/internal/usecase/ppp/manage_ppp.go): Orkestrasi bisnis untuk Secrets, Profiles, Active Sessions, dan Inactive subscribers.
  - [internal/usecase/ppp/manage_ppp_test.go](file:///home/quixiq/projects/work/polyglot/internal/usecase/ppp/manage_ppp_test.go): Unit test 100% passing.
- **ConnectRPC Adapter Layer**:
  - [internal/adapter/connect/ppp/](file:///home/quixiq/projects/work/polyglot/internal/adapter/connect/ppp/):
    - `ppp_handler.go`, `secret_handler.go`, `profile_handler.go`, `active_handler.go`, `inactive_handler.go`, `stream_handler.go`, `mapper.go`, `router.go`.
- **Security & Authorization**:
  - [internal/adapter/auth/procedure_permissions.go](file:///home/quixiq/projects/work/polyglot/internal/adapter/auth/procedure_permissions.go): Pemetaan procedure ke `ppp:read` dan `ppp:manage`.
  - [internal/adapter/auth/policy_seeder.go](file:///home/quixiq/projects/work/polyglot/internal/adapter/auth/policy_seeder.go): Default Casbin seeding untuk admin dan teknisi.
- **Application Routing**:
  - [internal/app/app.go](file:///home/quixiq/projects/work/polyglot/internal/app/app.go): Mount `PPPServiceHandler` ke protected mux.

---

### 3. Frontend Web (`web/src/features/ppp/`)
- **API Client & Hooks**:
  - [web/src/lib/api-client.ts](file:///home/quixiq/projects/work/polyglot/web/src/lib/api-client.ts): Export `pppClient`.
  - [web/src/features/ppp/api/](file:///home/quixiq/projects/work/polyglot/web/src/features/ppp/api/): React Query hooks (`usePPPSecretsQuery`, `usePPPProfilesQuery`, `usePPPActiveSessionsQuery`, `usePPPInactiveSecretsQuery`, mutations, dan live streaming).
- **Tabs & Tables**:
  - **Secrets Tab**:
    - Table dengan `DataTableToolbar`, faceted filters (**Profile**, **Service**, **Status**, **Comment/Batch**), quick search, dan sorting.
    - Password generator & show/hide toggle di modal form Add/Edit Secret.
    - Status toggle (Enable / Disable) per-row.
    - Floating Bulk Actions: Mass Enable, Mass Disable, dan Mass Delete.
  - **Active Sessions Tab**:
    - Real-time active connections table dengan info IP, MAC, Uptime, Traffic, dan Auth (RADIUS/Local).
    - Single session Disconnect / Kick action.
    - Floating Bulk Actions: Mass Disconnect untuk sesi yang dipilih.
    - Live stream toggle switch.
  - **Inactive Tab**:
    - Daftar pelanggan offline (perbandingan secrets vs active sessions) lengkap dengan faceted filters & quick actions.
  - **Profiles Tab**:
    - CRUD PPP Bandwidth & IP Profiles (Preset chips 2M/5M/10M/20M/50M/100M/Isolir, IP Pools, DNS push, Address Lists, Only-One).
- **Navigation & Routing**:
  - [web/src/routes/_authenticated/ppp/index.tsx](file:///home/quixiq/projects/work/polyglot/web/src/routes/_authenticated/ppp/index.tsx): Route TanStack Router `/_authenticated/ppp/`.
  - [web/src/components/layout/data/sidebar-data.ts](file:///home/quixiq/projects/work/polyglot/web/src/components/layout/data/sidebar-data.ts): Menu **PPPoE / PPP** di sidebar dengan icon `Network`.

---

## 🧪 Hasil Verifikasi

### 1. Backend Verification
```bash
go test -race ./...
```
**Output**: Seluruh package lulus 100% tanpa race condition (`ok`).

```bash
go build -v ./cmd/server
```
**Output**: Server binary ter-compile sukses tanpa error.

### 2. Frontend Web Verification
```bash
pnpm --prefix web build
```
**Output**:
- TypeScript typecheck lulus tanpa error (`tsc -b`).
- Bundle asset `dist/assets/ppp-Dw1ZCAqN.js` (59.79 kB) berhasil di-generate.
