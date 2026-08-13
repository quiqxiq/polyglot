import { readFileSync, writeFileSync } from 'node:fs';

const assembled = JSON.parse(readFileSync('.understand-anything/intermediate/assembled-graph.json', 'utf8'));
const validIds = new Set(assembled.nodes.map(n => n.id));

const fileNodes = assembled.nodes.filter(n =>
  ['file', 'config', 'document', 'service', 'resource', 'schema', 'table'].includes(n.type) && n.filePath
);

const byPrefix = (prefix) => fileNodes.filter(n => n.filePath.startsWith(prefix)).map(n => n.id);
const byExact = (p) => fileNodes.filter(n => n.filePath === p).map(n => n.id);
const byPath = (pred) => fileNodes.filter(n => pred(n.filePath)).map(n => n.id);

const layerDefs = [
  { id: 'layer:entrypoint', name: 'Entry Points & Commands', description: 'Executable entry points: the API server, seed tooling, and probe utility that bootstrap the application.', nodeIds: byPrefix('cmd/') },
  { id: 'layer:domain', name: 'Domain Layer', description: 'Pure business entities and rules — device, command, session, customer, subscription, billing, bot, knowledge, LLM, plan — with no I/O or framework imports.', nodeIds: byPrefix('internal/domain/') },
  { id: 'layer:usecase', name: 'Use Case Layer', description: 'Orchestrates single use cases (network command execution, device management, hotspot, bot engine, billing) depending only on domain and ports.', nodeIds: byPrefix('internal/usecase/') },
  { id: 'layer:port', name: 'Port Layer (Contracts)', description: 'Go interfaces defining contracts: DeviceDriver, TerminalDeviceDriver, repositories, credential vault, LLM provider, WhatsApp gateway, audit writer.', nodeIds: byPrefix('internal/port/') },
  { id: 'layer:adapter-connect', name: 'Adapters — ConnectRPC & HTTP & WS', description: 'Transport adapters: ConnectRPC handlers (auth, device, hotspot, billing, bot, customer), HTTP middleware, WebSocket terminal streaming, MCP tools.', nodeIds: [...byPrefix('internal/adapter/connect/'), ...byPrefix('internal/adapter/http/'), ...byPrefix('internal/adapter/ws/'), ...byPrefix('internal/adapter/mcp/'), ...byPrefix('internal/adapter/auth/')] },
  { id: 'layer:adapter-persistence', name: 'Adapters — Persistence & Infra', description: 'Postgres repositories and models, Redis store, AES credential vault, LLM provider clients (Claude, Gemini, OpenAI, Groq).', nodeIds: [...byPrefix('internal/adapter/postgres/'), ...byPrefix('internal/adapter/redis/'), ...byPrefix('internal/adapter/vault/'), ...byPrefix('internal/adapter/llm/')] },
  { id: 'layer:driver', name: 'Device Drivers', description: 'Vendor-specific network device drivers: MikroTik RouterOS (API via goros + dual-connection streaming), generic SSH/Telnet (scrapligo), Cisco, Huawei OLT, ZTE OLT, NETCONF, GenieACS, WhatsApp gateway.', nodeIds: byPrefix('internal/driver/') },
  { id: 'layer:platformdef', name: 'Platform Definitions', description: 'Vendor-as-data: YAML platform definitions (prompt patterns, login suffixes, command catalogs) consumed by the generic CLI engine.', nodeIds: byPrefix('internal/platformdef/') },
  { id: 'layer:core', name: 'Core Wiring & Cross-Cutting', description: 'Application composition root, driver registry, configuration/crypto, audit writer, and shared templates.', nodeIds: [...byPrefix('internal/app/'), ...byPrefix('internal/registry/'), ...byPrefix('internal/config/'), ...byPrefix('internal/audit/'), ...byPrefix('internal/templates/')] },
  { id: 'layer:pkg', name: 'Shared Utilities (pkg)', description: 'Reusable generic utilities independent of project domain: retry helper and voucher generator.', nodeIds: byPrefix('pkg/') },
  { id: 'layer:api-contracts', name: 'API Contracts & Generated Code', description: 'Protocol definitions (protobuf, OpenAPI, MCP tools) and generated ConnectRPC/gRPC/protobuf Go code.', nodeIds: byPrefix('api/') },
  { id: 'layer:data', name: 'Database Migrations', description: 'SQL schema migrations for devices, bot tables, and SSH port additions.', nodeIds: byPrefix('migrations/') },
  { id: 'layer:deployment', name: 'Deployment & Infrastructure', description: 'Dockerfile, docker-compose manifests, CI workflows, Makefile, and editor/agent configuration.', nodeIds: [...byPrefix('deployments/'), ...byExact('.github/workflows/ci.yml'), ...byExact('Makefile'), ...byExact('.golangci.yml'), ...byExact('.dockerignore'), ...byExact('.mcp.json'), ...byExact('opencode.jsonc'), ...byExact('.env.example'), ...byExact('go.mod'), ...byExact('go.sum')] },
  { id: 'layer:docs', name: 'Documentation', description: 'Project docs: architecture, ADRs, system structure, tech stack, MikroTik command reference, and agent instructions.', nodeIds: byPath(p => /\.md$/.test(p)) },
];

// First-assignment-wins dedup, preserving priority order
const assigned = new Set();
const layers = layerDefs.map(def => {
  const nodeIds = [...new Set(def.nodeIds.filter(id => validIds.has(id) && !assigned.has(id)))];
  nodeIds.forEach(id => assigned.add(id));
  return { ...def, nodeIds };
});

// Any remaining file nodes land in misc
const remaining = fileNodes.filter(n => !assigned.has(n.id)).map(n => n.id);
if (remaining.length) {
  layers.push({ id: 'layer:unassigned', name: 'Other / Misc', description: 'Files not otherwise categorized.', nodeIds: remaining });
}

writeFileSync('.understand-anything/intermediate/layers.json', JSON.stringify(layers, null, 2));
console.log('REBUILT layers.json —', layers.length, 'layers');
for (const l of layers) console.log(`  ${l.id}: ${l.nodeIds.length} nodes`);
const dupCheck = new Set();
let dups = 0;
for (const l of layers) for (const id of l.nodeIds) { if (dupCheck.has(id)) dups++; dupCheck.add(id); }
console.log('duplicate assignments:', dups);
