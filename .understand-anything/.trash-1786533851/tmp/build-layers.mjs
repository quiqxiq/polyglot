#!/usr/bin/env node
// Assigns every file-level node to an architectural layer based on its path,
// mirroring the previous graph's 13-layer structure.
import fs from 'fs';
import path from 'path';

const ROOT = '/home/quixiq/projects/polyground/polyglot';
const GRAPH_PATH = path.join(ROOT, '.understand-anything/intermediate/assembled-graph.json');
const graph = JSON.parse(fs.readFileSync(GRAPH_PATH, 'utf8'));

const fileLevelTypes = new Set(['file', 'config', 'document', 'service', 'pipeline', 'table', 'schema', 'endpoint']);
const fileNodes = graph.nodes.filter((n) => fileLevelTypes.has(n.type));

// Layer definitions: id, name, description, matcher
const LAYER_RULES = [
  {
    id: 'layer:entry-points',
    name: 'Entry Points',
    description: 'Process entry points that bootstrap the application (server, probe, seed).',
    match: (p) => /^cmd\//.test(p),
  },
  {
    id: 'layer:domain',
    name: 'Domain Layer',
    description: 'Pure business entities, value objects, and domain errors — no external dependencies.',
    match: (p) => /^internal\/domain\//.test(p),
  },
  {
    id: 'layer:usecase',
    name: 'Use Case Layer',
    description: 'Application orchestration: network use cases (execute command, stream) and business use cases (manage customers, subscriptions, billing).',
    match: (p) => /^internal\/usecase\//.test(p),
  },
  {
    id: 'layer:ports',
    name: 'Port Contracts',
    description: 'Interface contracts (repositories, drivers, gateways) that adapters implement.',
    match: (p) => /^internal\/port\//.test(p),
  },
  {
    id: 'layer:adapters',
    name: 'Adapter Layer',
    description: 'Concrete adapters: HTTP/ConnectRPC handlers, MCP tools, WebSocket, Postgres, Redis, auth, LLM.',
    match: (p) => /^internal\/adapter\//.test(p),
  },
  {
    id: 'layer:drivers',
    name: 'Device Drivers & Registry',
    description: 'Vendor device drivers (Mikrotik, Cisco, generic SSH/Telnet, OLTs, GenieACS) and the driver registry.',
    match: (p) => /^internal\/driver\//.test(p) || /^internal\/registry\//.test(p),
  },
  {
    id: 'layer:core-infrastructure',
    name: 'Core Infrastructure',
    description: 'Application assembly, config loading, audit logging, crypto, and platform definitions.',
    match: (p) => /^internal\/app\//.test(p) || /^internal\/config\//.test(p) || /^internal\/audit\//.test(p) || /^internal\/platformdef\//.test(p) || /^internal\/templates\//.test(p),
  },
  {
    id: 'layer:api-contracts',
    name: 'API Contracts',
    description: 'Protobuf schemas, generated ConnectRPC code, and OpenAPI specifications.',
    match: (p) => /^api\//.test(p),
  },
  {
    id: 'layer:data',
    name: 'Data Layer',
    description: 'SQL migrations defining the database schema.',
    match: (p) => /^migrations\//.test(p),
  },
  {
    id: 'layer:documentation',
    name: 'Documentation',
    description: 'Markdown docs: README, architecture, ADRs, and operational guides.',
    match: (p) => /\.md$/.test(p) || /^docs\//.test(p),
  },
  {
    id: 'layer:deployment',
    name: 'Deployment & CI',
    description: 'Docker images, Compose manifests, and CI/CD pipelines.',
    match: (p) => /^deployments\//.test(p) || /\.dockerignore$/.test(p) || /^\.github\/workflows\//.test(p),
  },
  {
    id: 'layer:configuration',
    name: 'Configuration',
    description: 'Project-level configuration files (env, linters, module manifests, MCP config).',
    match: (p) => /^\.env\.example$/.test(p) || /^\.golangci\.yml$/.test(p) || /^\.mcp\.json$/.test(p) || /^go\.mod$/.test(p) || /^go\.sum$/.test(p) || /^Makefile$/.test(p) || /^opencode\.jsonc$/.test(p) || /^\.dockerignore$/.test(p),
  },
  {
    id: 'layer:shared-libraries',
    name: 'Shared Libraries',
    description: 'Reusable generic packages (retry, voucher).',
    match: (p) => /^pkg\//.test(p),
  },
];

const layers = LAYER_RULES.map((r) => ({ id: r.id, name: r.name, description: r.description, match: r.match, nodeIds: [] }));
const byId = new Map(layers.map((l) => [l.id, l]));

const unassigned = [];
for (const n of fileNodes) {
  let placed = false;
  for (const l of layers) {
    if (l.match && l.match(n.filePath)) {
      l.nodeIds.push(n.id);
      placed = true;
      break;
    }
  }
  if (!placed) unassigned.push(n.id);
}

// Handle leftover .dockerignore (configuration) and misc
for (const id of unassigned) {
  if (id.startsWith('config:.dockerignore')) byId.get('layer:configuration').nodeIds.push(id);
  else byId.get('layer:configuration').nodeIds.push(id);
}

// Clean layer arrays (remove the match fn)
const out = layers.map(({ match, ...l }) => ({ ...l }));
fs.writeFileSync(path.join(ROOT, '.understand-anything/intermediate/layers.json'), JSON.stringify(out, null, 2));

const totalAssigned = out.reduce((a, l) => a + l.nodeIds.length, 0);
console.log('Layers built:', out.length);
console.log('File-level nodes assigned:', totalAssigned, 'of', fileNodes.length);
out.forEach((l) => console.log(`  ${l.id}: ${l.nodeIds.length}`));
