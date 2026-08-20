#!/usr/bin/env node
// Builds a guided tour that walks the user from entry point through the
// architecture layers, mirroring the previous graph's 13-step tour.
import fs from 'fs';
import path from 'path';

const ROOT = '/home/quixiq/projects/polyground/polyglot';
const GRAPH_PATH = path.join(ROOT, '.understand-anything/intermediate/assembled-graph.json');
const LAYERS_PATH = path.join(ROOT, '.understand-anything/intermediate/layers.json');
const graph = JSON.parse(fs.readFileSync(GRAPH_PATH, 'utf8'));
const layers = JSON.parse(fs.readFileSync(LAYERS_PATH, 'utf8'));

const layerById = new Map(layers.map((l) => [l.id, l]));

// Pick representative nodes per layer for the tour
function pickNodes(layerId, n = 3) {
  const layer = layerById.get(layerId);
  if (!layer) return [];
  return layer.nodeIds.slice(0, n);
}

const tour = [
  {
    order: 1,
    title: 'Project Overview',
    description: 'Polyglot is a NetOps + ISP management backend exposing MCP, REST, and WebSocket/SSE interfaces for multi-vendor network automation (Mikrotik, Cisco, generic SSH/Telnet, OLT vendors).',
    nodeIds: ['document:README.md', 'document:Polyglot-Architecture.md'],
  },
  {
    order: 2,
    title: 'Application Entry Points',
    description: 'The process starts here: the server binary boots configuration, database, Redis, auth, device drivers, and ConnectRPC services; probe runs connectivity checks; seed populates initial data.',
    nodeIds: pickNodes('layer:entry-points', 3),
  },
  {
    order: 3,
    title: 'Core Infrastructure — App Assembly',
    description: 'internal/app wires every layer together: dependencies, routers, and lifecycle management.',
    nodeIds: pickNodes('layer:core-infrastructure', 2),
  },
  {
    order: 4,
    title: 'Configuration & Crypto',
    description: 'Environment-driven configuration with validation, plus AES credential vault crypto helpers.',
    nodeIds: ['file:internal/config/config.go', 'file:internal/config/crypto.go', 'config:.env.example'],
  },
  {
    order: 5,
    title: 'Domain Layer — Business Entities',
    description: 'Pure domain entities and rules with zero external dependencies: device, command, session, customer, subscription, billing.',
    nodeIds: pickNodes('layer:domain', 4),
  },
  {
    order: 6,
    title: 'Port Contracts',
    description: 'Interfaces that decouple use cases from implementations: device drivers, repositories, credential vault, audit writer, LLM provider, WhatsApp gateway.',
    nodeIds: pickNodes('layer:ports', 3),
  },
  {
    order: 7,
    title: 'Use Case Layer — Orchestration',
    description: 'Application logic: execute commands, stream device output, manage customers/subscriptions/billing, hotspot operations.',
    nodeIds: pickNodes('layer:usecase', 4),
  },
  {
    order: 8,
    title: 'Adapter Layer — HTTP/ConnectRPC & MCP',
    description: 'Concrete protocol adapters: ConnectRPC services, MCP tools, WebSocket handlers, Postgres repositories, Redis cache, auth (JWT + Casbin RBAC).',
    nodeIds: pickNodes('layer:adapters', 4),
  },
  {
    order: 9,
    title: 'Device Drivers & Registry',
    description: 'Vendor-specific drivers translate abstract operations into native commands: Mikrotik, Cisco, generic SSH/CLI, OLT vendors, GenieACS. The registry routes devices to the right driver.',
    nodeIds: pickNodes('layer:drivers', 4),
  },
  {
    order: 10,
    title: 'API Contracts — Protobuf & ConnectRPC',
    description: 'Service contracts defined in protobuf and exposed via ConnectRPC: device, hotspot, customer, billing, bot, knowledge, RBAC, WhatsApp.',
    nodeIds: ['schema:api/proto/v1/device.proto', 'schema:api/proto/v1/hotspot.proto', 'schema:api/proto/v1/whatsapp.proto'],
  },
  {
    order: 11,
    title: 'Data Layer — Migrations',
    description: 'SQL migrations define the schema: devices, users, conversations, vouchers, knowledge, technicians, and more.',
    nodeIds: ['table:migrations/000001_create_devices_table.up.sql', 'table:migrations/000002_create_bot_tables.up.sql'],
  },
  {
    order: 12,
    title: 'Shared Libraries & Docs',
    description: 'Reusable generic packages (retry, voucher generation) and the architecture/ADR documentation.',
    nodeIds: pickNodes('layer:shared-libraries', 2),
  },
  {
    order: 13,
    title: 'Deployment & CI',
    description: 'Docker image, Compose manifests, and GitHub Actions CI that build and ship the server.',
    nodeIds: pickNodes('layer:deployment', 3),
  },
];

fs.writeFileSync(path.join(ROOT, '.understand-anything/intermediate/tour.json'), JSON.stringify(tour, null, 2));
console.log('Tour built with', tour.length, 'steps.');
