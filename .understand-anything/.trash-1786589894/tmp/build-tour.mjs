import { readFileSync, writeFileSync } from 'node:fs';

const assembled = JSON.parse(readFileSync('.understand-anything/intermediate/assembled-graph.json', 'utf8'));
const validIds = new Set(assembled.nodes.map(n => n.id));
const fileNode = (p) => {
  const id = `file:${p}`;
  return validIds.has(id) ? id : null;
};

const steps = [
  {
    order: 1,
    title: 'Project Overview',
    description: 'NetOps + ISP management backend in Go. Start with README and architecture docs to understand the clean architecture (domain/usecase/port/adapter/driver) and the multi-vendor network automation vision.',
    nodeIds: ['document:README.md', 'document:Polyglot-Architecture.md', 'document:SYSTEM-STRUCTURE-AND-ARCHITECTURE.md'].filter(id => validIds.has(id)),
  },
  {
    order: 2,
    title: 'Architecture Decision Records',
    description: 'Key decisions: Gin over Echo (0001), device driver without separate session (0002), MikroTik dual-connection streaming (0003), and generic CLI driver based on scrapligo (0004).',
    nodeIds: ['document:docs/adr/0001-pilih-gin-daripada-echo.md', 'document:docs/adr/0002-devicedriver-tanpa-session-terpisah.md', 'document:docs/adr/0003-mikrotik-dual-connection-streaming.md', 'document:docs/adr/0004-generic-cli-driver-scrapligo.md'].filter(id => validIds.has(id)),
  },
  {
    order: 3,
    title: 'Server Entry Point',
    description: 'cmd/server/main.go boots the HTTP server. cmd/seed/main.go seeds users, Casbin policies, knowledge, and LLM config; cmd/probe/main.go probes services.',
    nodeIds: ['file:cmd/server/main.go', 'file:cmd/seed/main.go', 'file:cmd/probe/main.go'].filter(Boolean),
  },
  {
    order: 4,
    title: 'Application Composition Root',
    description: 'internal/app/app.go wires everything: repositories, vault, LLM providers, driver registry, and all use cases and handlers.',
    nodeIds: [fileNode('internal/app/app.go')].filter(Boolean),
  },
  {
    order: 5,
    title: 'Configuration & Security',
    description: 'internal/config holds environment config, AES credential encryption (crypto.go), Casbin RBAC model, and the system prompt.',
    nodeIds: ['file:internal/config/config.go', 'file:internal/config/crypto.go', 'file:internal/config/rbac_model.conf'].filter(Boolean),
  },
  {
    order: 6,
    title: 'Domain Layer',
    description: 'Pure business entities: device, command (with operation enums), session, customer, subscription, billing invoice, bot conversation, knowledge, LLM config.',
    nodeIds: ['file:internal/domain/device/device.go', 'file:internal/domain/command/command.go', 'file:internal/domain/command/policy.go', 'file:internal/domain/device/credentials.go'].filter(Boolean),
  },
  {
    order: 7,
    title: 'Port Layer — Contracts',
    description: 'Go interfaces decoupling use cases from implementations: DeviceDriver, TerminalDeviceDriver, StreamingDeviceDriver, repositories, CredentialVault, LLMProvider, WhatsAppGateway, ValidatingDriver.',
    nodeIds: ['file:internal/port/device_driver.go', 'file:internal/port/streaming_driver.go', 'file:internal/port/validating_driver.go', 'file:internal/port/credential_vault.go'].filter(Boolean),
  },
  {
    order: 8,
    title: 'Driver Registry',
    description: 'internal/registry maps vendor/transport to concrete drivers and dispatches commands, including dual-connection MikroTik exec/stream handling.',
    nodeIds: ['file:internal/registry/registry.go'].filter(Boolean),
  },
  {
    order: 9,
    title: 'MikroTik Driver — RouterOS API (goros)',
    description: 'The heart of the project: driver.go (dual-connection exec/stream), connect.go (RouterOS API dial via goros), commands.go (command catalog + Classify/Translate), system.go, ip.go, hotspot, ppp, queue, firewall helpers, and stream.go (output streaming).',
    nodeIds: ['file:internal/driver/mikrotik/driver.go', 'file:internal/driver/mikrotik/connect.go', 'file:internal/driver/mikrotik/commands.go', 'file:internal/driver/mikrotik/stream.go', 'file:internal/driver/mikrotik/system.go'].filter(Boolean),
  },
  {
    order: 10,
    title: 'Generic CLI Engine (SSH/Telnet via scrapligo)',
    description: 'genericcli session engine + genericssh/generictelnet drivers use platform YAML + Catalog as vendor-as-data; genericssh also implements TerminalDeviceDriver for interactive PTY sessions (pty.go).',
    nodeIds: ['file:internal/driver/genericcli/session.go', 'file:internal/driver/genericcli/catalog.go', 'file:internal/driver/genericssh/driver.go', 'file:internal/driver/genericssh/pty.go', 'file:internal/driver/generictelnet/driver.go', 'file:internal/platformdef/mikrotik_routeros.yaml'].filter(Boolean),
  },
  {
    order: 11,
    title: 'Other Device Drivers',
    description: 'Cisco, Huawei OLT, ZTE OLT (SNMP/telnet), NETCONF, and GenieACS (TR-069) drivers plus the WhatsApp gateway driver.',
    nodeIds: ['file:internal/driver/cisco/driver.go', 'file:internal/driver/huaweiolt/driver.go', 'file:internal/driver/zteolt/telnet.go', 'file:internal/driver/netconf/driver.go', 'file:internal/driver/genieacs/client.go', 'file:internal/driver/whatsapp/client.go'].filter(Boolean),
  },
  {
    order: 12,
    title: 'Network Use Cases',
    description: 'ExecuteCommand (with goros Validate gates), GetDeviceStatus, PushConfig, OpenTerminal/StreamTerminal (interactive PTY), GetActiveSessions, StreamOutput.',
    nodeIds: ['file:internal/usecase/network/execute_command.go', 'file:internal/usecase/network/get_device_status.go', 'file:internal/usecase/network/open_terminal.go', 'file:internal/usecase/network/stream_terminal.go', 'file:internal/usecase/network/get_active_sessions.go'].filter(Boolean),
  },
  {
    order: 13,
    title: 'Business & Bot Use Cases',
    description: 'Device management CRUD, hotspot operations (voucher, profile, active users), bot engine with guardrails, knowledge retrieval, customer/subscription/billing management.',
    nodeIds: ['file:internal/usecase/device/manage_device.go', 'file:internal/usecase/hotspot/hotspot_usecase.go', 'file:internal/usecase/bot/engine.go', 'file:internal/usecase/knowledge/retriever.go', 'file:internal/usecase/billing/manage_invoice.go'].filter(Boolean),
  },
  {
    order: 14,
    title: 'ConnectRPC Handlers (device)',
    description: 'The main device handler exposes ConnectRPC services for device CRUD, command execution, status, probe, and terminal streaming (bidi).',
    nodeIds: ['file:internal/adapter/connect/device/device_handler.go', 'file:internal/adapter/connect/device/probe_handler.go'].filter(Boolean),
  },
  {
    order: 15,
    title: 'WebSocket & SSE Streaming',
    description: 'WS adapter bridges PTY stdout to browser xterm.js; SSE hub streams command output for live reboot and other long-running ops.',
    nodeIds: ['file:internal/adapter/ws/device_stream_handler.go', 'file:internal/adapter/ws/sse_hub.go', 'file:internal/adapter/ws/router.go'].filter(Boolean),
  },
  {
    order: 16,
    title: 'MCP Tools & LLM Adapters',
    description: 'Model Context Protocol tools (get_device_status, run_command, push_config, knowledge search, Mikhmon) and LLM provider adapters (Claude, Gemini, OpenAI, Groq).',
    nodeIds: ['file:internal/adapter/mcp/server.go', 'file:internal/adapter/mcp/tool_run_command.go', 'file:internal/adapter/mcp/tool_get_device_status.go', 'file:internal/adapter/llm/provider.go', 'file:internal/adapter/llm/claude/claude.go'].filter(Boolean),
  },
  {
    order: 17,
    title: 'Persistence Layer',
    description: 'Postgres repositories + GORM models for devices, users, bot conversations, knowledge, LLM config; Redis cache; AES vault.',
    nodeIds: ['file:internal/adapter/postgres/store.go', 'file:internal/adapter/postgres/device_repository.go', 'file:internal/adapter/postgres/models/device_model.go', 'file:internal/adapter/redis/store.go', 'file:internal/adapter/vault/aes_vault.go'].filter(Boolean),
  },
  {
    order: 18,
    title: 'API Contracts & Migrations',
    description: 'Protobuf definitions for ConnectRPC services (auth, device, hotspot, billing, bot, knowledge, probe, rbac, whatsapp) and SQL migrations.',
    nodeIds: ['file:api/proto/v1/device.proto', 'file:api/proto/v1/hotspot.proto', 'file:api/openapi.yaml', 'file:migrations/000001_create_devices_table.up.sql'].filter(Boolean),
  },
  {
    order: 19,
    title: 'Deployment',
    description: 'Dockerfile and docker-compose for the backend plus dependencies; CI workflow and Makefile for build/lint/test.',
    nodeIds: ['file:deployments/docker/Dockerfile', 'file:deployments/docker-compose.yml', 'file:.github/workflows/ci.yml', 'file:Makefile'].filter(Boolean),
  },
];

for (const s of steps) s.nodeIds = s.nodeIds.filter(id => validIds.has(id));

writeFileSync('.understand-anything/intermediate/tour.json', JSON.stringify(steps, null, 2));
console.log('WROTE tour.json —', steps.length, 'steps');
for (const s of steps) console.log(`  ${s.order}. ${s.title} (${s.nodeIds.length} nodes)`);
