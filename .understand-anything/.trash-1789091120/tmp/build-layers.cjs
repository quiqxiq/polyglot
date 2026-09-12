#!/usr/bin/env node
/**
 * build-layers.cjs — Phase 4: deterministic architectural layering.
 * Every file-level node is assigned to exactly one layer (Hexagonal/Onion,
 * innermost first). Mirrors AGENTS.md repository shape.
 */
const fs = require('fs');
const path = require('path');
const ROOT = process.argv[2];
const GPATH = path.join(ROOT, '.understand-anything', 'intermediate', 'assembled-graph.json');
const g = JSON.parse(fs.readFileSync(GPATH, 'utf8'));

const FILE_LEVEL = new Set(['file', 'config', 'document', 'service', 'pipeline', 'table', 'schema', 'resource', 'endpoint']);
const nodes = g.nodes.filter(n => FILE_LEVEL.has(n.type));
const nodeIds = new Set(nodes.map(n => n.id));

const layers = [
  {
    id: 'layer:entrypoint',
    name: 'Titik Masuk (Entrypoint)',
    description: 'Program executable: server utama (HTTP+ConnectRPC+MCP+WS/SSE), probe, dan seeder database.',
    match: (p, t) => /^cmd\//.test(p) || p === 'Makefile' || p === '.air.toml',
  },
  {
    id: 'layer:contract',
    name: 'Kontrak API (Protobuf & OpenAPI)',
    description: 'Definisi kontrak transport: skema protobuf (layanan ConnectRPC + endpoint rpc) dan spesifikasi OpenAPI HTTP.',
    match: (p, t) => t === 'schema' || t === 'endpoint' || p === 'api/openapi.yaml' || p === 'buf.yaml' || p === 'buf.gen.yaml',
  },
  {
    id: 'layer:http-adapter',
    name: 'Adapter Transport (ConnectRPC, HTTP, WS/SSE, MCP)',
    description: 'Handler transport dan middleware: ConnectRPC per-context, HTTP mux, WebSocket/SSE streaming, dan MCP server.',
    match: (p) => /^internal\/adapter\/(connect|http|ws|mcp)\//.test(p),
  },
  {
    id: 'layer:infra-adapter',
    name: 'Adapter Infrastruktur (DB, Redis, WhatsApp, LLM, Vendor)',
    description: 'Implementasi konkrit port: PostgreSQL/GORM, Redis, WhatsApp (whatsmeow), LLM (Genkit/OpenAI), storage, provisioner, auth Casbin/JWT.',
    match: (p) => /^internal\/adapter\/(postgres|redis|whatsapp|llm|storage|provisioner|tripay|auth)\//.test(p),
  },
  {
    id: 'layer:driver',
    name: 'Driver Perangkat (Vendor CLI/API)',
    description: 'Driver eksternal multi-vendor: generic CLI via Scrapli (MikroTik RouterOS dsb.) dan implementasi protokol perangkat jaringan.',
    match: (p) => /^internal\/driver\//.test(p) || /^internal\/platformdef\//.test(p),
  },
  {
    id: 'layer:usecase',
    name: 'Use Case (Orkestrasi Bisnis)',
    description: 'Orkestrasi alur bisnis di atas domain dan port: auth, billing, bot, hotspot, ppp, device, notification, portal, dll.',
    match: (p) => /^internal\/usecase\//.test(p),
  },
  {
    id: 'layer:domain',
    name: 'Domain (Model & Invariant)',
    description: 'Model domain murni tanpa dependensi framework: entitas, value object, invariant, dan error domain per-context (user, billing, bot, device, …).',
    match: (p) => /^internal\/domain\//.test(p),
  },
  {
    id: 'layer:port',
    name: 'Port (Kontrak Interface)',
    description: 'Interface consumer-owned antara use case dan adapter/driver — batas depensi heksagonal.',
    match: (p) => /^internal\/port\//.test(p),
  },
  {
    id: 'layer:shared-kernel',
    name: 'Shared Kernel (pkg/ & Config)',
    description: 'Paket lintas-context: logger, response, fault, crypto, validator, plus konfigurasi aplikasi (internal/config) dan registry/app wiring.',
    match: (p) => /^pkg\//.test(p) || /^internal\/(config|registry|app)\//.test(p),
  },
  {
    id: 'layer:cross-cutting',
    name: 'Aset & Lintas-Batas (Template, Voucher, Data)',
    description: 'Aset template struk voucher (embed.FS), generator voucher, definisi skill bot, dan aset data statis.',
    match: (p) => /^internal\/(template|voucher)\//.test(p) || /^data\//.test(p),
  },
  {
    id: 'layer:database',
    name: 'Skema Database (Migrasi SQL)',
    description: 'Evolusi skema PostgreSQL/TimescaleDB via golang-migrate: tabel perangkat, bot, billing, ISP core, ping metrics.',
    match: (p, t) => t === 'table' && /^migrations\//.test(p),
  },
  {
    id: 'layer:tests',
    name: 'Test Integrasi',
    description: 'Uji integrasi bertag (PostgreSQL/Redis/testcontainers): driver MikroTik, streaming, Timescale, migrasi, WhatsApp.',
    match: (p) => /^test\//.test(p),
  },
  {
    id: 'layer:deploy',
    name: 'Deployment & Infrastruktur',
    description: 'Docker Compose dev/prod, Dockerfile, Nginx reverse proxy, dan konfigurasi build/lint proyek.',
    match: (p, t) => t === 'service' || /^deployments\//.test(p) || /^\.golangci\.yml$/.test(p) || /^\.dockerignore$/.test(p),
  },
  {
    id: 'layer:docs',
    name: 'Dokumentasi',
    description: 'README, pedoman pengembangan, arsitektur, ADR, spesifikasi fitur (mikhmon, database), dan rencana kerja.',
    match: (p, t) => t === 'document' || /^docs\//.test(p),
  },
  {
    id: 'layer:tooling',
    name: 'Tooling & Skrip',
    description: 'Skrip smoke-test API (ConnectRPC/SSE), skrip boundary check, dan konfigurasi tooling (MCP, opencode).',
    match: (p, t) => /^scripts\//.test(p) || /^\.mcp\.json$/.test(p) || /^opencode\.jsonc$/.test(p) || p === '.env.example' || p === 'go.mod' || p === 'go.sum',
  },
];

const assigned = new Map();
const layersOut = [];
for (const L of layers) {
  layersOut.push({ id: L.id, name: L.name, description: L.description, nodeIds: [] });
}
const layerById = new Map(layersOut.map(l => [l.id, l]));

let unassigned = [];
for (const n of nodes) {
  const p = n.filePath || '';
  const t = n.type;
  let hit = null;
  for (const L of layers) {
    if (L.match(p, t)) { hit = L.id; break; }
  }
  if (!hit) { unassigned.push(`${t}:${p}`); continue; }
  if (assigned.has(n.id)) { unassigned.push(`${t}:${p} (dup)`); continue; }
  assigned.set(n.id, hit);
  layerById.get(hit).nodeIds.push(n.id);
}

// every file-level node must be assigned exactly once
fs.writeFileSync(path.join(ROOT, '.understand-anything', 'intermediate', 'layers.json'), JSON.stringify(layersOut, null, 1));
console.log(JSON.stringify({
  layers: layersOut.map(l => ({ id: l.id, count: l.nodeIds.length })),
  assignedTotal: assigned.size,
  fileLevelNodes: nodes.length,
  unassigned: unassigned.length,
  unassignedSample: unassigned.slice(0, 15),
}, null, 1));
