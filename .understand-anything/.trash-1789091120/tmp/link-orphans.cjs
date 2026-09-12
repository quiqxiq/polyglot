#!/usr/bin/env node
/**
 * link-orphans.cjs — Phase 3 fix-up: graph-wide linking pass.
 * Adds deterministic edges for non-code files (docs, configs, migrations,
 * infra, scripts) that got no edges because per-batch scoping missed
 * cross-batch targets. Drops .understand-anything/ analysis artifacts.
 */
const fs = require('fs');
const path = require('path');

const ROOT = process.argv[2];
const GPATH = path.join(ROOT, '.understand-anything', 'intermediate', 'assembled-graph.json');
const g = JSON.parse(fs.readFileSync(GPATH, 'utf8'));

// ---- 0. drop analysis artifacts -------------------------------------------
const before = g.nodes.length;
g.nodes = g.nodes.filter(n => !(n.filePath && n.filePath.startsWith('.understand-anything/')));
const droppedArtifacts = before - g.nodes.length;
if (droppedArtifacts) {
  const keep = new Set(g.nodes.map(n => n.id));
  g.edges = g.edges.filter(e => keep.has(e.source) && keep.has(e.target));
}

const nodeIds = new Set(g.nodes.map(n => n.id));
const fileNodes = g.nodes.filter(n => n.type === 'file');

function firstNode(pred) {
  const n = g.nodes.find(pred);
  return n ? n.id : null;
}
function findFile(re) {
  return firstNode(n => n.type === 'file' && re.test(n.filePath));
}
const edgeKey = (e) => `${e.source}|${e.target}|${e.type}`;
const existing = new Set(g.edges.map(edgeKey));
function addEdge(source, target, type, weight) {
  if (!source || !target || source === target) return false;
  if (!nodeIds.has(source) || !nodeIds.has(target)) return false;
  const e = { source, target, type, weight };
  if (existing.has(edgeKey(e))) return false;
  existing.add(edgeKey(e));
  g.edges.push(e);
  return true;
}

// ---- helpers ---------------------------------------------------------------
const REP = {};
Object.assign(REP, {
  main: findFile(/cmd\/server\/main\.go/),
  configLoader: findFile(/internal\/config\/config\.go/),
  casbin: findFile(/internal\/adapter\/auth\/casbin\.go/),
  voucherGen: findFile(/internal\/voucher\/generator\.go/),
  templateHandler: findFile(/internal\/adapter\/connect\/hotspot\/template_handler\.go/),
  catalog: findFile(/internal\/driver\/genericcli\/catalog\.go/),
  migSmoke: findFile(/internal\/adapter\/postgres\/migrations_smoke_test\.go/),
  hotspotRep: findFile(/internal\/adapter\/connect\/hotspot\/[a-z_]+_handler\.go/) || findFile(/internal\/usecase\/hotspot\//),
  httpRep: findFile(/internal\/adapter\/http\//),
  pkgRep: findFile(/pkg\/logger\/logger\.go/),
  pingRep: findFile(/internal\/usecase\/ping\//) || findFile(/internal\/usecase\/monitoring\//),
  chatwootRep: findFile(/internal\/adapter\/connect\/bot\//) || findFile(/internal\/usecase\/bot\//),
});
Object.assign(REP, {
  skillsRep: findFile(/internal\/usecase\/skills\//) || findFile(/internal\/usecase\/bot\//) || findFile(/internal\/usecase\/knowledge\//) || REP.main,
  waRep: findFile(/internal\/adapter\/connect\/whatsapp\//) || findFile(/internal\/adapter\/wa\//) || REP.main,
  ispCore: firstNode(n => n.type === 'table' && /000015_rebuild_isp_core\.up\.sql/.test(n.id)),
  billingCore: firstNode(n => n.type === 'table' && /000006_create_billing_tables\.up\.sql/.test(n.id)),
  devicesTable: firstNode(n => n.type === 'table' && /000001_create_devices_table\.up\.sql/.test(n.id)),
});

// ---- 1. migrations ----------------------------------------------------------
// up → rollback pair (related), up → smoke test (tested_by)
const byNumber = {};
for (const n of g.nodes) {
  const m = n.id.match(/^table:migrations\/(\d+)_[^.]+\.(up|down)\.sql$/);
  if (m && n.type === 'table') {
    byNumber[m[1]] = byNumber[m[1]] || {};
    byNumber[m[1]][m[2]] = n.id;
  }
}
let migEdges = 0;
for (const [num, pair] of Object.entries(byNumber)) {
  if (pair.up && pair.down) {
    if (addEdge(pair.up, pair.down, 'related', 0.5)) migEdges++;
  }
  if (pair.up && REP.migSmoke) {
    if (addEdge(pair.up, REP.migSmoke, 'tested_by', 0.5)) migEdges++;
  }
}

// ---- 2. embedded assets & platform definitions ------------------------------
let assetEdges = 0;
const templateTxt = g.nodes.filter(n => n.type === 'document' && /^internal\/template\/.*\.txt$/.test(n.filePath || ''));
for (const t of templateTxt) {
  for (const consumer of [REP.voucherGen, REP.templateHandler]) {
    if (addEdge(`document:${t.filePath}`, consumer, 'related', 0.5)) assetEdges++;
  }
}
if (addEdge('config:internal/config/rbac_model.conf', REP.configLoader, 'configures', 0.6)) assetEdges++;
if (addEdge('config:internal/platformdef/mikrotik_routeros.yaml', REP.catalog, 'configures', 0.6)) assetEdges++;
if (addEdge('document:internal/platformdef/README.md', 'config:internal/platformdef/mikrotik_routeros.yaml', 'documents', 0.5)) assetEdges++;

// integration fixtures → integration suite representative
const fixtures = g.nodes.filter(n => /^test\/integration\/fixtures\//.test(n.filePath || ''));
const intRep = findFile(/test\/integration\/[a-z_]+_test\.go/);
for (const f of fixtures) {
  if (addEdge(`${f.type}:${f.filePath}`, intRep, 'related', 0.5)) assetEdges++;
}

// ---- 3. infra / root configs ------------------------------------------------
let infraEdges = 0;
const serviceOrphans = g.nodes.filter(n => n.type === 'service');
for (const s of serviceOrphans) {
  const p = s.filePath || '';
  let target = REP.main;
  if (/nginx/.test(p)) target = REP.main;
  if (/prometheus|grafana|otel/.test(p)) target = REP.pingRep || REP.main;
  if (addEdge(`service:${p}`, target, /Dockerfile/.test(p) ? 'deploys' : /nginx/.test(p) ? 'serves' : 'configures', /Dockerfile/.test(p) ? 0.7 : 0.5)) infraEdges++;
}
const rootConfigs = {
  'config:go.mod': ['config:Makefile', 'related'],
  'config:go.sum': ['config:go.mod', 'related'],
  'config:.env.example': [REP.configLoader, 'related'],
  'config:.golangci.yml': ['config:Makefile', 'related'],
  'config:.air.toml': [REP.main, 'related'],
  'config:.dockerignore': ['service:deployments/docker/Dockerfile', 'related'],
  'config:buf.yaml': [firstNode(n => n.type === 'schema'), 'configures'],
  'config:buf.gen.yaml': [firstNode(n => n.type === 'schema'), 'configures'],
  'config:api/openapi.yaml': [REP.httpRep, 'documents'],
  'config:Makefile': [REP.main, 'configures'],
  'config:.mcp.json': [REP.main, 'configures'],
  'config:opencode.jsonc': ['config:Makefile', 'related'],
};
for (const [src, [target, type]] of Object.entries(rootConfigs)) {
  if (addEdge(src, target, type, 0.6)) infraEdges++;
}
// data/skills definitions → bot skills usecase
for (const n of g.nodes.filter(n => /^data\/skills\//.test(n.filePath || ''))) {
  if (addEdge(`config:${n.filePath}`, REP.skillsRep, 'configures', 0.6)) infraEdges++;
}
if (addEdge('document:data/system-prompt.md', REP.chatwootRep || REP.skillsRep, 'configures', 0.6)) infraEdges++;

// ---- 4. scripts → server ----------------------------------------------------
let scriptEdges = 0;
for (const n of g.nodes.filter(n => /^scripts\//.test(n.filePath || '') && n.type === 'config')) {
  if (addEdge(`config:${n.filePath}`, REP.main, 'depends_on', 0.5)) scriptEdges++;
}

// ---- 5. documents ------------------------------------------------------------
let docEdges = 0;
const docsToTargets = [
  [/^README\.md$/, [REP.main, 'documents']],
  [/^AGENTS\.md$/, [REP.main, 'documents']],
  [/^CLAUDE\.md$/, [REP.main, 'documents']],
  [/^DEVELOPMENT-GUIDELINES\.md$/, [REP.main, 'documents']],
  [/^EFFECTIVE_GO\.md$/, [REP.pkgRep, 'documents']],
  [/^Polyglot-Architecture\.md$/, [REP.main, 'documents']],
  [/^SYSTEM-STRUCTURE-AND-ARCHITECTURE\.md$/, [REP.main, 'documents']],
  [/^PLAN\.md$/, [REP.main, 'documents']],
  [/^PLAN-BACKEND-STANDARDIZATION-V2\.md$/, [REP.main, 'documents']],
  [/^PLAN-PING-ANALYTICS-OPTIMIZATION\.md$/, [REP.pingRep, 'documents']],
  [/^ROADMAP-AND-ISSUES\.md$/, [REP.main, 'documents']],
  [/^CHATWOOT_PLAN\.md$/, [REP.chatwootRep, 'documents']],
  [/^provision-plan\.md$/, [REP.main, 'documents']],
  [/^TECH-STACK-DAN-PERSIAPAN\.md$/, [REP.main, 'documents']],
  [/^docs\/database-schema\.md$/, [REP.ispCore || REP.billingCore, 'documents']],
  [/^docs\/DATABASE-SCHEMA-ISP\.md$/, [REP.billingCore || REP.ispCore, 'documents']],
  [/^docs\/BACKEND-MIGRATION-ROADMAP\.md$/, [REP.main, 'documents']],
];
for (const n of g.nodes.filter(n => n.type === 'document')) {
  const p = n.filePath || '';
  for (const [re, [target, type]] of docsToTargets) {
    if (re.test(p)) { if (addEdge(`document:${p}`, target, type, 0.5)) docEdges++; break; }
  }
  if (/^docs\/mikhmon\//.test(p)) {
    if (addEdge(`document:${p}`, REP.hotspotRep, 'documents', 0.5)) docEdges++;
    continue;
  }
  if (/^docs\/adr\//.test(p)) {
    // keyword match ADR slug → code file
    const slug = p.toLowerCase();
    const keywords = ['genericcli', 'scrapligo', 'ping', 'timescale', 'billing', 'whatsmeow', 'whatsapp', 'mcp', 'sse', 'hotspot', 'mikhmon', 'casbin', 'rbac', 'connect'];
    let matched = null;
    for (const kw of keywords) {
      if (slug.includes(kw)) {
        matched = findFile(new RegExp(kw)) || null;
        if (matched) break;
      }
    }
    if (addEdge(`document:${p}`, matched || REP.main, 'documents', 0.5)) docEdges++;
    continue;
  }
  if (/^docs\/backend\//.test(p)) {
    if (addEdge(`document:${p}`, REP.main, 'documents', 0.5)) docEdges++;
  }
}

// ---- 6. stragglers: Makefile is typed 'service', nginx is config, skills docs are documents
let fixEdges = 0;
const MAKEFILE = firstNode(n => n.filePath === 'Makefile');
const pingDocsTarget = findFile(/internal\/adapter\/connect\/device\/metrics_handler\.go/);
const pingModelTarget = findFile(/internal\/adapter\/postgres\/model\/ping_metric_model\.go/);
const stragglers = [
  ['config:.golangci.yml', MAKEFILE, 'related'],
  ['config:opencode.jsonc', MAKEFILE, 'related'],
  ['config:deployments/nginx/conf.d/default.conf', REP.main, 'serves'],
  ['config:deployments/nginx/nginx.conf', REP.main, 'serves'],
  ['document:PLAN-PING-ANALYTICS-OPTIMIZATION.md', pingDocsTarget, 'documents'],
  ['document:PLAN-PING-ANALYTICS-OPTIMIZATION.md', pingModelTarget, 'documents'],
  ['config:test/integration/.gitkeep', intRep, 'related'],
];
for (const [src, target, type] of stragglers) {
  if (addEdge(src, target, type, 0.5)) fixEdges++;
}
// data/skills skill documents → skills usecase (documents)
for (const n of g.nodes.filter(n => /^data\/skills\//.test(n.filePath || '') && n.type === 'document')) {
  if (addEdge(`document:${n.filePath}`, REP.skillsRep, 'documents', 0.5)) fixEdges++;
}

fs.writeFileSync(GPATH, JSON.stringify(g, null, 1));
const withEdges = new Set();
g.edges.forEach(e => { withEdges.add(e.source); withEdges.add(e.target); });
const orphans = g.nodes.filter(n => !withEdges.has(n.id)).length;
console.log(JSON.stringify({
  droppedArtifacts,
  added: { migrations: migEdges, assets: assetEdges, infra: infraEdges, scripts: scriptEdges, docs: docEdges, fixes: fixEdges },
  totalNodes: g.nodes.length,
  totalEdges: g.edges.length,
  remainingOrphans: orphans,
  missingReps: Object.entries(REP).filter(([, v]) => !v).map(([k]) => k),
}));
