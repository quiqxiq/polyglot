#!/usr/bin/env node
// Converts deterministic tree-sitter extraction results into the
// knowledge-graph batch format consumed by merge-batch-graphs.py.
// Reproduces the structure of the previous graph build:
//  - code files -> file: nodes (with function/class child nodes)
//  - functions/classes kept when exported OR span >= 8
//  - non-code files -> config:/document:/service:/pipeline:/schema:/endpoint:/table:
//  - edges: contains, imports, exports, calls, documents, configures,
//    deploys, depends_on, triggers
import fs from 'fs';
import path from 'path';

const ROOT = '/home/quixiq/projects/polyground/polyglot';
const INTER = path.join(ROOT, '.understand-anything/intermediate');
const TMP = path.join(ROOT, '.understand-anything/tmp');

// ---- Load scan + batches ----
const scan = JSON.parse(fs.readFileSync(path.join(INTER, 'scan-result.json'), 'utf8'));
const batches = JSON.parse(fs.readFileSync(path.join(INTER, 'batches.json'), 'utf8'));

const fileMeta = {};
for (const f of scan.files) fileMeta[f.path] = f;
const importMap = scan.importMap || {};

// ---- Load all extraction results into a per-file lookup ----
const extractionByFile = {};
const FUNCTION_IDS = new Map(); // "path|name" -> id
const CLASS_IDS = new Map();
for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    extractionByFile[r.path] = r;
    for (const fn of r.functions || []) FUNCTION_IDS.set(`${r.path}|${fn.name}`, `function:${r.path}:${fn.name}`);
    for (const c of r.classes || []) CLASS_IDS.set(`${r.path}|${c.name}`, `class:${r.path}:${c.name}`);
  }
}

const complexityFromLines = (n) => {
  if (n > 600) return 'complex';
  if (n > 200) return 'moderate';
  return 'simple';
};

// ---- Node ID classification ----
// Mirrors the previous build: code files are file:, everything else gets
// a semantic prefix.
function classifyNodeId(rel) {
  const f = fileMeta[rel];
  const cat = f ? f.fileCategory : 'code';
  const lower = rel.toLowerCase();

  // infra
  if (cat === 'infra') {
    if (/\.dockerignore$/.test(rel) || /dockerfile/.test(lower) || /docker-compose/.test(lower)) return { prefix: 'service', type: 'service' };
    if (/\.github\/workflows\//.test(rel)) return { prefix: 'pipeline', type: 'pipeline' };
    return { prefix: 'service', type: 'service' };
  }
  // data: migrations -> table, proto -> schema, openapi -> endpoint
  if (cat === 'data') {
    if (/migrations\/.*\.sql$/.test(rel)) return { prefix: 'table', type: 'table' };
    if (/\.proto$/.test(rel)) return { prefix: 'schema', type: 'schema' };
    if (/openapi/.test(lower)) return { prefix: 'endpoint', type: 'endpoint' };
    return { prefix: 'table', type: 'table' };
  }
  if (cat === 'config') {
    if (/Makefile$/.test(rel)) return { prefix: 'config', type: 'config' };
    return { prefix: 'config', type: 'config' };
  }
  if (cat === 'docs') return { prefix: 'document', type: 'document' };
  if (cat === 'markup' || cat === 'script') return { prefix: 'file', type: 'file' };
  return { prefix: 'file', type: 'file' };
}

// ---- Summary generation ----
function fileSummary(rel, ext) {
  const exports = ext ? [...(ext.functions || []), ...(ext.classes || [])].filter((s) => /^[A-Z]/.test(s.name)).map((s) => s.name).slice(0, 6) : [];
  const dir = path.dirname(rel);
  let summary = `${rel} — ${fileMeta[rel]?.fileCategory || 'code'} file in ${dir}`;
  if (exports.length) summary += `, exports: ${exports.join(', ')}`;
  if (ext?.callGraph?.length) summary += `, ${ext.callGraph.length} internal call sites`;
  return summary;
}

// ---- Build nodes/edges for one batch ----
function buildBatch(batch) {
  const idx = batch.batchIndex;
  const nodes = [];
  const edges = [];
  const batchFiles = batch.files.map((f) => f.path);

  for (const rel of batchFiles) {
    const ext = extractionByFile[rel];
    const { prefix, type } = classifyNodeId(rel);
    const f = fileMeta[rel] || { language: 'unknown', sizeLines: 0 };

    if (type === 'file') {
      // ---- code file node ----
      const fileId = `file:${rel}`;
      nodes.push({
        id: fileId,
        type: 'file',
        name: path.basename(rel),
        filePath: rel,
        summary: fileSummary(rel, ext),
        tags: ['go', 'code'],
        complexity: complexityFromLines(f.sizeLines || 0),
      });

      // function + class nodes (exported OR span >= 8)
      const fnNodes = [];
      for (const fn of ext?.functions || []) {
        const span = (fn.endLine || 0) - (fn.startLine || 0);
        if (!/^[A-Z]/.test(fn.name) && span < 8) continue;
        const id = `function:${rel}:${fn.name}`;
        const exported = /^[A-Z]/.test(fn.name);
        fnNodes.push({
          id,
          type: 'function',
          name: fn.name,
          filePath: rel,
          lineRange: [fn.startLine, fn.endLine],
          summary: `${fn.name}() in ${path.basename(rel)}${exported ? ' (exported)' : ''}`,
          tags: exported ? ['function', 'exported'] : ['function'],
          complexity: span > 80 ? 'moderate' : 'simple',
        });
      }
      const clsNodes = [];
      for (const c of ext?.classes || []) {
        const span = (c.endLine || 0) - (c.startLine || 0);
        const inTest = /_test\.go$/.test(rel);
        if (!/^[A-Z]/.test(c.name) && span < 8 && !inTest) continue;
        const id = `class:${rel}:${c.name}`;
        const exported = /^[A-Z]/.test(c.name);
        clsNodes.push({
          id,
          type: 'class',
          name: c.name,
          filePath: rel,
          lineRange: c.startLine ? [c.startLine, c.endLine || c.startLine] : undefined,
          summary: `${c.name} type in ${path.basename(rel)}${exported ? ' (exported)' : ''}`,
          tags: exported ? ['class', 'exported'] : ['class'],
          complexity: 'simple',
        });
      }
      nodes.push(...fnNodes, ...clsNodes);

      // contains edges
      for (const n of [...fnNodes, ...clsNodes]) edges.push({ source: fileId, target: n.id, type: 'contains', direction: 'forward', weight: 1.0 });
      // exports edges (exported symbols only)
      for (const n of [...fnNodes, ...clsNodes]) {
        if (/^[A-Z]/.test(n.name)) edges.push({ source: fileId, target: n.id, type: 'exports', direction: 'forward', weight: 0.8 });
      }
      // calls edges (internal callees)
      for (const call of ext?.callGraph || []) {
        const target = FUNCTION_IDS.get(`${rel}|${call.callee}`);
        if (target) edges.push({ source: fileId, target, type: 'calls', direction: 'forward', weight: 0.8 });
      }
      // imports edges
      for (const imp of importMap[rel] || []) {
        if (fileMeta[imp] && imp !== rel) {
          edges.push({ source: fileId, target: `file:${imp}`, type: 'imports', direction: 'forward', weight: 0.7 });
        }
      }
    } else {
      // ---- non-code file node ----
      const fileId = `${prefix}:${rel}`;
      let summary = `${rel} — ${type} definition`;
      if (type === 'schema') summary = `Protobuf schema defining ${path.basename(rel).replace('.proto', '')} service contract — messages and RPC endpoints`;
      if (type === 'endpoint' && /openapi/.test(rel)) summary = 'OpenAPI specification for the polyglot REST API';
      if (type === 'service' && /Dockerfile/.test(rel)) summary = 'Dockerfile building the polyglot server image';
      if (type === 'service' && /docker-compose/.test(rel)) summary = 'Docker Compose deployment manifest';
      if (type === 'table') summary = `SQL migration ${path.basename(rel)}`;
      if (type === 'pipeline') summary = 'GitHub Actions CI pipeline';
      if (type === 'document') summary = `Documentation: ${path.basename(rel)}`;
      if (type === 'config') summary = `Configuration file ${path.basename(rel)}`;

      nodes.push({
        id: fileId,
        type,
        name: path.basename(rel),
        filePath: rel,
        summary,
        tags: [type, fileMeta[rel]?.language || 'config'],
        complexity: complexityFromLines(f.sizeLines || 0),
      });

      // ---- child nodes for non-code files ----
      if (type === 'schema' && ext) {
        for (const ep of ext.endpoints || []) {
          const epName = ep.path || ep.name || 'rpc';
          const epId = `endpoint:${rel}:${epName}`;
          nodes.push({
            id: epId,
            type: 'endpoint',
            name: epName,
            filePath: rel,
            lineRange: ep.startLine ? [ep.startLine, ep.endLine || ep.startLine] : undefined,
            summary: `RPC endpoint ${epName} exposed by the ${path.basename(rel).replace('.proto', '')} service`,
            tags: ['rpc', 'endpoint', 'protobuf'],
            complexity: 'simple',
          });
          edges.push({ source: fileId, target: epId, type: 'contains', direction: 'forward', weight: 1.0 });
        }
      }
      if (type === 'service' && ext && ext.services?.length) {
        // Dockerfile stages / compose services
        for (const svc of ext.services) {
          const childId = `${prefix}:${rel}:${svc.name}`;
          nodes.push({
            id: childId,
            type: 'service',
            name: svc.name,
            filePath: rel,
            lineRange: svc.startLine ? [svc.startLine, svc.endLine || svc.startLine] : undefined,
            summary: /Dockerfile/.test(rel)
              ? `Docker build stage ${svc.name} in the server image`
              : `Service ${svc.name} in the compose manifest`,
            tags: ['docker', 'build-stage', 'containerization'],
            complexity: 'simple',
          });
          edges.push({ source: fileId, target: childId, type: 'contains', direction: 'forward', weight: 1.0 });
        }
      }
      if (type === 'table' && ext && /\.up\.sql$/.test(rel)) {
        for (const def of ext.definitions || []) {
          if (def.kind !== 'table') continue;
          const tblId = `table:${rel}:${def.name}`;
          nodes.push({
            id: tblId,
            type: 'table',
            name: def.name,
            filePath: rel,
            lineRange: def.startLine ? [def.startLine, def.endLine || def.startLine] : undefined,
            summary: `Database table ${def.name} with columns: ${(def.fields || []).join(', ')}.`,
            tags: ['database', 'table', 'schema'],
            complexity: 'simple',
          });
          edges.push({ source: fileId, target: tblId, type: 'contains', direction: 'forward', weight: 1.0 });
        }
      }

      // ---- cross-file edges ----
      if (type === 'document') {
        // docs -> primary entry point + app assembly
        if (/README|CLAUDE|AGENTS|SYSTEM-STRUCTURE|PANDUAN/.test(rel)) {
          edges.push({ source: fileId, target: 'file:cmd/server/main.go', type: 'documents', direction: 'forward', weight: 0.7 });
        }
        if (/Architecture|BACKEND-MIGRATION|TECH-STACK/.test(rel)) {
          edges.push({ source: fileId, target: 'file:internal/app/app.go', type: 'documents', direction: 'forward', weight: 0.7 });
        }
      }
      if (type === 'config') {
        if (/\.env\.example/.test(rel)) {
          edges.push({ source: fileId, target: 'file:cmd/server/main.go', type: 'configures', direction: 'forward', weight: 0.6 });
          edges.push({ source: fileId, target: 'file:internal/config/config.go', type: 'configures', direction: 'forward', weight: 0.6 });
        }
        if (/go\.mod/.test(rel)) {
          edges.push({ source: fileId, target: 'file:cmd/server/main.go', type: 'depends_on', direction: 'forward', weight: 0.6 });
        }
      }
      if (type === 'service') {
        if (/Dockerfile/.test(rel)) {
          edges.push({ source: fileId, target: 'file:cmd/server/main.go', type: 'deploys', direction: 'forward', weight: 0.7 });
        }
        if (/docker-compose/.test(rel)) {
          edges.push({ source: fileId, target: 'service:deployments/docker/Dockerfile', type: 'depends_on', direction: 'forward', weight: 0.6 });
        }
      }
      if (type === 'pipeline') {
        edges.push({ source: fileId, target: 'file:cmd/server/main.go', type: 'triggers', direction: 'forward', weight: 0.6 });
        edges.push({ source: fileId, target: 'config:.golangci.yml', type: 'triggers', direction: 'forward', weight: 0.6 });
      }
      if (type === 'endpoint' && /openapi/.test(rel)) {
        edges.push({ source: fileId, target: 'file:cmd/server/main.go', type: 'documents', direction: 'forward', weight: 0.7 });
      }
    }
  }

  return { nodes, edges };
}

// ---- Iterate batches ----
let totalNodes = 0;
let totalEdges = 0;
const writtenBatches = [];

for (const batch of batches.batches) {
  const idx = batch.batchIndex;
  const { nodes, edges } = buildBatch(batch);
  fs.writeFileSync(path.join(INTER, `batch-${idx}.json`), JSON.stringify({ nodes, edges }, null, 1));
  writtenBatches.push(idx);
  totalNodes += nodes.length;
  totalEdges += edges.length;
}

console.log(`Wrote ${writtenBatches.length} batch files (${totalNodes} nodes, ${totalEdges} edges total).`);
