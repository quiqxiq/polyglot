#!/usr/bin/env node
/**
 * build-tour.cjs — Phase 5: guided tour (Bahasa Indonesia).
 * Follows one request from entrypoint to database and back, plus testing & ops.
 */
const fs = require('fs');
const path = require('path');
const ROOT = process.argv[2];
const GPATH = path.join(ROOT, '.understand-anything', 'intermediate', 'assembled-graph.json');
const g = JSON.parse(fs.readFileSync(GPATH, 'utf8'));
const ids = new Set(g.nodes.map(n => n.id));
const nodeOf = (id) => g.nodes.find(n => n.id === id);

const FILE = 'file:';
const CFG = 'config:';
const DOC = 'document:';
const SVC = 'service:';
const TBL = 'table:';
const SCH = 'schema:';
const EP = 'endpoint:';

const steps = [
  {
    order: 1,
    title: 'Peta Besar: Engine NetOps & ISP Polyglot',
    description: 'Mulailah dari README dan dokumentasi arsitektur. Polyglot adalah engine backend Go murni (net/http.ServeMux, Go 1.26) untuk otomasi jaringan multi-vendor, billing ISP, hotspot voucher, dan bot AI WhatsApp — mengekspos ConnectRPC, MCP, serta SSE/WebSocket. Arsitektur mengikuti Clean Hexagonal: domain di tengah, port sebagai kontrak, adapter/driver di tepi. Dokumen AGENTS.md adalah sumber aturan non-negotiable (batas layer, error handling, logging).',
    nodeIds: [DOC + 'README.md', DOC + 'AGENTS.md', DOC + 'Polyglot-Architecture.md', DOC + 'SYSTEM-STRUCTURE-AND-ARCHITECTURE.md'].filter(id => ids.has(id)),
  },
  {
    order: 2,
    title: 'Bootstrap: main.go, App, dan Router',
    description: 'Alur program dimulai di cmd/server/main.go: memuat konfigurasi (internal/config), menyiapkan logger terpusat (pkg/logger), membuka PostgreSQL dan Redis, lalu menyusun dependency injection melalui registry (internal/registry) dan wiring di internal/app/app.go. router.go memasang mux: /connect/* untuk ConnectRPC, jalur MCP, serta streaming WS/SSE. scheduler.go menjalankan pekerjaan periodik (polling perangkat, notifikasi) dengan ownership goroutine yang jelas.',
    nodeIds: [FILE + 'cmd/server/main.go', FILE + 'internal/config/config.go', FILE + 'internal/registry/registry.go', FILE + 'internal/app/app.go', FILE + 'internal/app/router.go', FILE + 'internal/app/scheduler.go', FILE + 'pkg/logger/logger.go'].filter(id => ids.has(id)),
  },
  {
    order: 3,
    title: 'Kontrak API: Protobuf & ConnectRPC',
    description: 'Setiap layanan didefinisikan sebagai skema protobuf di api/proto/v1/ (auth, billing, device, hotspot, ppp, whatsapp, dst.). Endpoint rpc (240 total) menjadi kontrak yang di-handoff ke handler ConnectRPC. Kode generated (api/gen/) tidak pernah diedit manual — buf lint + make proto-check menjaga konsistensi dan kompatibilitas wire. Konfigurasi buf.yaml/buf.gen.yaml mengatur lint, breaking check, dan generasi.',
    nodeIds: [SCH + 'api/proto/v1/auth.proto', SCH + 'api/proto/v1/billing.proto', SCH + 'api/proto/v1/device.proto', SCH + 'api/proto/v1/hotspot.proto', SCH + 'api/proto/v1/ppp.proto'].filter(id => ids.has(id)).concat([
      'endpoint:api/proto/v1/auth.proto:AuthService.Login', 'endpoint:api/proto/v1/billing.proto:BillingService.*',
    ].map(p => p.includes('*') ? null : p).filter(id => id && ids.has(id))),
  },
  {
    order: 4,
    title: 'Adapter Transport: Handler ConnectRPC, HTTP, WS/SSE, MCP',
    description: 'Handler di internal/adapter/connect/<context>/ menerima permintaan RPC, memvalidasi, memanggil use case, lalu memetakan error melalui pkg/response — tidak ada connect.NewError ad hoc. Context besar: auth, billing, device, hotspot, ppp, whatsapp, portal, dll. internal/adapter/http menyediakan endpoint HTTP murni dan middleware (Chain), internal/adapter/ws menangani streaming terminal & log realtime (hub.go), dan internal/adapter/mcp mengekspos kemampuan engine ke klien MCP.',
    nodeIds: [FILE + 'internal/adapter/connect/provider.go', FILE + 'internal/adapter/connect/codec.go', FILE + 'internal/adapter/http/middleware/chain.go', FILE + 'internal/adapter/ws/hub.go'].filter(id => ids.has(id)).concat(g.nodes.filter(n => n.type === 'file' && /^internal\/adapter\/connect\/(auth|device|hotspot)\/[a-z_]+_handler\.go$/.test(n.filePath)).slice(0, 3).map(n => n.id)),
  },
  {
    order: 5,
    title: 'Use Case: Orkestrasi Bisnis',
    description: 'Use case di internal/usecase/<context>/ mengorkestrasi domain dan port tanpa tahu transport maupun vendor: auth (login, RBAC), billing (tagihan & pembayaran), hotspot (voucher Mikhmon-parity), ppp (secrets, profile, sesi aktif/nonaktif), device (ping metrics & stream), bot & llm (AI CS WhatsApp). Pola konstruktor dengan dependency injection membuat semuanya mudah diuji.',
    nodeIds: g.nodes.filter(n => n.type === 'file' && ['internal/usecase/auth/login.go', 'internal/usecase/auth/refresh_token.go', 'internal/usecase/billing/run_billing.go', 'internal/usecase/billing/checkout.go', 'internal/usecase/hotspot/hotspot_usecase.go', 'internal/usecase/hotspot/manage_voucher.go', 'internal/usecase/ppp/manage_ppp.go', 'internal/usecase/device/device_usecase.go'].includes(n.filePath)).slice(0, 6).map(n => n.id),
  },
  {
    order: 6,
    title: 'Domain & Port: Jantung Heksagonal',
    description: 'internal/domain/<context>/ berisi model murni: entitas (User, Invoice, PPPSecret, Device), value object, invariant, dan error domain dibuat dengan fault.New — tanpa import transport, database, atau framework. internal/port/ mendeklarasikan kontrak yang dibutuhkan use case: repository (customer_repository.go, cashbook_repository.go), device_driver.go untuk vendor, credential_vault.go, authorizer.go (Casbin RBAC). Ketergantungan selalu mengarah keluar: domain ← usecase ← port ← adapter.',
    nodeIds: g.nodes.filter(n => n.type === 'file' && ['internal/domain/user/user.go', 'internal/domain/billing/invoice.go', 'internal/domain/ppp/ppp.go', 'internal/domain/device/device.go', 'internal/port/device_driver.go', 'internal/port/authorizer.go', 'internal/port/customer_repository.go'].includes(n.filePath)).slice(0, 8).map(n => n.id),
  },
  {
    order: 7,
    title: 'Driver & Adapter Infrastruktur: Vendor, DB, WhatsApp',
    description: 'internal/driver/ mengimplementasikan port device_driver untuk vendor nyata: mikrotik (API+SSH), genericcli/genericssh/generictelnet via Scrapli (didefinisikan deklaratif di platformdef/mikrotik_routeros.yaml), genieacs (TR-069), huaweiolt/zteolt (OLT), netconf, dan cisco. Sisi data: internal/adapter/postgres (GORM+store) dan redis untuk cache. WhatsApp berjalan di whatsmeow (adapter/whatsapp), LLM via Genkit/OpenAI (adapter/llm), pembayaran via adapter/tripay.',
    nodeIds: [FILE + 'internal/platformdef/mikrotik_routeros.yaml'].map(p => p.endsWith('.yaml') ? 'config:' + p : p).filter(id => ids.has(id)).concat(g.nodes.filter(n => n.type === 'file' && ['internal/driver/mikrotik/driver.go', 'internal/driver/mikrotik/connect.go', 'internal/driver/mikrotik/gateway.go', 'internal/driver/genericcli/catalog.go', 'internal/driver/genericcli/session.go', 'internal/driver/genieacs/driver.go', 'internal/adapter/postgres/store.go', 'internal/adapter/whatsapp/sender_adapter.go', 'internal/adapter/llm/provider.go'].includes(n.filePath)).slice(0, 7).map(n => n.id)),
  },
  {
    order: 8,
    title: 'Skema Database: Migrasi & Tabel Inti ISP',
    description: 'Skema berevolusi lewat golang-migrate (migrations/): 000001 devices, 000006 billing, 000015 rebuild ISP core (pelanggan, langganan, paket), 000022 device_ping_metrics (TimescaleDB hypertable), 000024 optimisasi ping metrics. Setiap migrasi berpasangan up/down dan diuji oleh migrations_smoke_test.go. Lihat docs/database-schema.md untuk peta domain data ISP.',
    nodeIds: [TBL + 'migrations/000001_create_devices_table.up.sql', TBL + 'migrations/000006_create_billing_tables.up.sql', TBL + 'migrations/000015_rebuild_isp_core.up.sql', TBL + 'migrations/000022_create_device_ping_metrics_table.up.sql', FILE + 'internal/adapter/postgres/migrations_smoke_test.go', DOC + 'docs/database-schema.md'].filter(id => ids.has(id)),
  },
  {
    order: 9,
    title: 'Kualitas, Test Integrasi, dan Operasional',
    description: 'Gerbang kualitas dijalankan lewat Makefile: make check (build+vet+test+lint+boundary checks), make proto-check, make test-integration (PostgreSQL/Redis/testcontainers: driver MikroTik, streaming, Timescale, migrasi), make security. Docker Compose dev/prod dan Nginx menutup sisi operasional. Skrip smoke-test di scripts/ menguji endpoint ConnectRPC dan SSE secara end-to-end. Untuk kontribusi baru, ikuti DEVELOPMENT-GUIDELINES.md dan EFFECTIVE_GO.md.',
    nodeIds: [SVC + 'Makefile', FILE + 'internal/adapter/postgres/migrations_smoke_test.go', DOC + 'DEVELOPMENT-GUIDELINES.md', DOC + 'EFFECTIVE_GO.md', SVC + 'deployments/docker-compose.yml', CFG + 'deployments/nginx/nginx.conf'].filter(id => ids.has(id)).concat(g.nodes.filter(n => n.type === 'file' && /^test\/integration\/.*_test\.go$/.test(n.filePath)).slice(0, 4).map(n => n.id)),
  },
];

// Drop steps whose nodeIds ended up empty; ensure every nodeId exists (already filtered above)
const out = steps.filter(s => s.nodeIds.length > 0).map((s, i) => ({ ...s, order: i + 1 }));
fs.writeFileSync(path.join(ROOT, '.understand-anything', 'intermediate', 'tour.json'), JSON.stringify(out, null, 1));
console.log(JSON.stringify(out.map(s => ({ order: s.order, title: s.title, nodes: s.nodeIds.length })), null, 1));
