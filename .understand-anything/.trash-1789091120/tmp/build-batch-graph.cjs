#!/usr/bin/env node
/**
 * build-batch-graph.cjs — Phase 2 inline analyzer for /understand
 *
 * Replaces the LLM file-analyzer subagent (unavailable in this runtime) with a
 * deterministic converter over the bundled tree-sitter extractor:
 *   extract-structure.mjs (Go functions/classes/callGraph, proto defs, SQL tables)
 * + import-map.json (pre-resolved project-internal imports)
 *
 * Produces one intermediate/batch-<i>.json per batch in batches.json,
 * in the GraphNode/GraphEdge format expected by merge-batch-graphs.py.
 *
 * Usage: node build-batch-graph.cjs <projectRoot> <skillDir> [batchIndex...]
 */
const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');

const PROJECT_ROOT = process.argv[2];
const SKILL_DIR = process.argv[3];
const ONLY_BATCHES = process.argv.slice(4).map(Number);
const LANG_NAME = 'Bahasa Indonesia';

const INTER = path.join(PROJECT_ROOT, '.understand-anything', 'intermediate');
const TMP = path.join(PROJECT_ROOT, '.understand-anything', 'tmp');

const batches = JSON.parse(fs.readFileSync(path.join(INTER, 'batches.json'), 'utf8')).batches;
const importMap = JSON.parse(fs.readFileSync(path.join(INTER, 'import-map.json'), 'utf8')).importMap || {};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
const stripExt = (p) => p.replace(/\.[^.]+$/, '');

function goTypeNode(file, cls) {
  return { id: `class:${file}:${cls.name}`, type: 'class', name: cls.name, filePath: file };
}
function goFuncNode(file, fn) {
  return { id: `function:${file}:${fn.name}`, type: 'function', name: fn.name, filePath: file };
}

function tagify(...parts) {
  const t = [...new Set(parts.filter(Boolean).map(s => String(s).toLowerCase().replace(/[^a-z0-9-]+/g, '-').replace(/^-+|-+$/g, '')))];
  return t.length ? t : ['untagged'];
}

function complexity(lines) {
  if (lines > 800) return 'complex';
  if (lines > 300) return 'moderate';
  return 'simple';
}

// Layer role inference from path — used in summaries/tags (backend Go hexagonal)
function roleOf(file) {
  const seg = file.split('/');
  if (file.startsWith('cmd/')) return { tag: 'entrypoint', id: 'Titik masuk aplikasi' };
  if (seg[1] === 'domain') return { tag: 'domain', id: 'Model domain murni: entitas, invariant, dan error domain' };
  if (seg[1] === 'usecase') return { tag: 'usecase', id: 'Orkestrasi use case di atas domain dan port' };
  if (seg[1] === 'port') return { tag: 'port', id: 'Kontrak interface milik consumer (port hexagonal)' };
  if (seg[1] === 'adapter' && seg[2] === 'connect') return { tag: 'connectrpc', id: 'Handler ConnectRPC: router, handler, dan mapper protobuf-domain' };
  if (seg[1] === 'adapter' && seg[2] === 'http') return { tag: 'http', id: 'Handler HTTP dan middleware (net/http.ServeMux)' };
  if (seg[1] === 'adapter') return { tag: 'adapter', id: 'Adapter infrastruktur: implementasi konkrit dari port' };
  if (seg[1] === 'driver') return { tag: 'driver', id: 'Driver eksternal: implementasi protokol/hardware vendor' };
  if (seg[1] === 'handler' || seg[1] === 'router') return { tag: 'http', id: 'Handler HTTP dan routing' };
  if (file.startsWith('pkg/')) return { tag: 'pkg', id: 'Paket shared library lintas-context' };
  if (file.startsWith('test/')) return { tag: 'test', id: 'Uji integrasi dengan dependensi layanan eksplisit' };
  return { tag: 'go', id: 'Berkas kode Go proyek' };
}

// ---------------------------------------------------------------------------
// Per-category node/edge builders
// ---------------------------------------------------------------------------

function buildGoFile(res, imp) {
  const file = res.path;
  const lines = res.totalLines || 0;
  const role = roleOf(file);
  const pkgName = path.basename(path.dirname(file));

  const fn = res.functions || [];
  const cls = res.classes || [];
  const cg = res.callGraph || [];
  const localFnNames = new Set(fn.map(f => f.name));

  const summary = `${role.id} — paket \`${pkgName}\` (${lines} baris, ${fn.length} fungsi, ${cls.length} tipe).`;

  const descParts = [];
  if (cls.length) {
    descParts.push('Tipe utama: ' + cls.slice(0, 5).map(c => c.name).join(', ') + (cls.length > 5 ? ', …' : '') + '.');
  }
  if (fn.length) {
    const top = fn.slice().sort((a, b) => (b.endLine - b.startLine) - (a.endLine - a.startLine)).slice(0, 5);
    descParts.push('Fungsi terbesar: ' + top.map(f => f.name).join(', ') + '.');
  }
  if (imp && imp.length) {
    descParts.push(`Bergantung pada ${imp.length} berkas internal.`);
  }
  const description = (summary + ' ' + descParts.join(' ')).trim();

  const tags = tagify('go', role.tag, pkgName, lines > 500 ? 'besar' : 'inti');

  const nodes = [{ id: `file:${file}`, type: 'file', name: path.basename(file), filePath: file, language: 'go', complexity: complexity(lines), summary, description, tags }];

  // Only export significant symbols to keep the graph manageable:
  // all classes/structs + exported functions (Go: uppercase-first) with size >= 25 lines.
  for (const c of cls) {
    const s = `Tipe ${c.name} di ${path.basename(file)}` + (c.methods && c.methods.length ? `, dengan method: ${c.methods.slice(0, 6).join(', ')}` : '') + '.';
    nodes.push({ ...goTypeNode(file, c), complexity: 'simple', summary: s, description: s + ` Didefinisikan pada baris ${c.startLine}-${c.endLine}.`, tags: tagify('go', 'tipe', c.name.toLowerCase()) });
  }
  for (const f of fn) {
    const isExported = /^[A-Z]/.test(f.name);
    const size = (f.endLine || f.startLine) - (f.startLine || 0);
    if (!isExported && size < 25) continue;
    const s = `Fungsi ${f.name} (${size} baris)` + (f.params && f.params.length ? `, parameter: ${f.params.slice(0, 5).join(', ')}` : '') + '.';
    nodes.push({ ...goFuncNode(file, f), complexity: size > 150 ? 'complex' : size > 60 ? 'moderate' : 'simple', summary: s, description: `${s} Didefinisikan pada baris ${f.startLine}-${f.endLine}.`, tags: tagify('go', 'fungsi', f.name.toLowerCase()) });
  }

  const edges = [];
  edges.push({ source: `file:${file}`, target: path.basename(file), type: 'contains', weight: 1.0, _replaceTargetPath: file });
  for (const c of cls) edges.push({ source: `file:${file}`, target: `class:${file}:${c.name}`, type: 'contains', weight: 1.0 });
  for (const f of fn) edges.push({ source: `file:${file}`, target: `function:${file}:${f.name}`, type: 'contains', weight: 1.0 });
  // imports (from pre-resolved import map)
  for (const target of (imp || [])) {
    if (target === file) continue;
    edges.push({ source: `file:${file}`, target: `file:${target}`, type: 'imports', weight: 0.7 });
  }
  // intra-file call graph (bound to local function nodes when possible)
  const seen = new Set();
  for (const call of cg) {
    const srcNode = localFnNames.has(call.caller) ? `function:${file}:${call.caller}` : `file:${file}`;
    // callee is usually external (stdlib/vendor) — only bind when it resolves to a local symbol
    if (!localFnNames.has(call.callee)) continue;
    const key = `${srcNode}->function:${file}:${call.callee}`;
    if (seen.has(key)) continue;
    seen.add(key);
    edges.push({ source: srcNode, target: `function:${file}:${call.callee}`, type: 'calls', weight: 0.8 });
  }
  return { nodes, edges };
}

function buildProtoFile(res) {
  const file = res.path;
  const lines = res.totalLines || 0;
  const defs = res.definitions || [];
  const eps = res.endpoints || [];
  const base = path.basename(file);
  const services = [...new Set((eps || []).map(e => String(e.path || '').split('.')[0]).filter(Boolean))];

  const summary = `Skema protobuf ${base} — ${defs.length} definisi (${new Set(defs.map(d => d.kind)).size} jenis) dan ${eps.length} rpc.`;
  const description = `${summary} Definisi utama: ${defs.slice(0, 6).map(d => d.name).join(', ')}${defs.length > 6 ? ', …' : ''}.${services.length ? ` Layanan: ${services.join(', ')}.` : ''} Nama field, nomor, dan semantik wire bersifat backward-compatible.`;
  const tags = tagify('protobuf', 'skema', services.length ? services.map(s => s.replace(/Service$/, '').toLowerCase()) : []);

  const nodes = [{ id: `schema:${file}`, type: 'schema', name: base, filePath: file, complexity: complexity(lines), summary, description, tags }];
  const edges = [];
  for (const e of eps) {
    const svc = String(e.path || '').split('.')[0];
    const name = String(e.path || base).split('.').slice(1).join('.') || base;
    nodes.push({ id: `endpoint:${file}:${e.path}`, type: 'endpoint', name: e.path, filePath: file, complexity: 'simple', summary: `RPC ${e.path} (proto).`, description: `Endpoint rpc ${e.path} pada baris ${e.startLine} dari ${base}. Ditangani oleh handler ConnectRPC yang sesuai di internal/adapter/connect.`, tags: tagify('rpc', svc ? svc.replace(/Service$/, '').toLowerCase() : '', name.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase()) });
    edges.push({ source: `schema:${file}`, target: `endpoint:${file}:${e.path}`, type: 'defines_schema', weight: 0.8 });
  }
  // message definitions that are request/response of endpoints → related (keep light: skip to avoid noise)
  return { nodes, edges };
}

function buildSqlFile(res) {
  const file = res.path;
  const lines = res.totalLines || 0;
  const defs = res.definitions || [];
  const isDown = file.endsWith('.down.sql');
  const tables = defs.filter(d => d.kind === 'table');
  const others = defs.filter(d => d.kind !== 'table');
  const summary = (isDown ? `Migrasi rollback — menghapus/perubahan mundur skema (${tables.length} tabel terdampak).` : `Migrasi skema — membuat/mengubah ${tables.length} tabel.`) + ` (${lines} baris)`;
  const description = `${summary}${tables.length ? ' Tabel: ' + tables.map(t => t.name).join(', ') + '.' : ''}${others.length ? ` Objek lain: ${others.slice(0, 5).map(o => o.name).join(', ')}${others.length > 5 ? ', …' : ''}.` : ''}`;
  const tags = tagify('sql', 'migrasi', isDown ? 'rollback' : 'naik');

  const nodes = [{ id: `table:${file}`, type: 'table', name: path.basename(file), filePath: file, complexity: complexity(lines), summary, description, tags }];
  const edges = [];
  // Represent each real table of a .up migration as its own node? Keep the file node
  // as the unit (merge normalizes); per-table nodes for up-migrations:
  if (!isDown) {
    for (const t of tables.slice(0, 12)) {
      nodes.push({ id: `table:${file}:${t.name}`, type: 'table', name: t.name, filePath: file, complexity: 'simple', summary: `Tabel \`${t.name}\` — ${t.fields.length} kolom.`, description: `Tabel \`${t.name}\` dibuat/diubah oleh ${path.basename(file)} (baris ${t.startLine}-${t.endLine}). Kolom: ${t.fields.slice(0, 12).join(', ')}${t.fields.length > 12 ? ', …' : ''}.`, tags: tagify('sql', 'tabel', t.name.replace(/_/g, '-')) });
      edges.push({ source: `table:${file}`, target: `table:${file}:${t.name}`, type: 'contains', weight: 1.0 });
    }
  }
  return { nodes, edges };
}

function buildNonCodeNode(f) {
  const file = f.path;
  const cat = f.fileCategory;
  const base = path.basename(file);
  const lines = f.sizeLines || 0;
  let type = 'config', summary, tags;
  if (cat === 'docs' || /\.md$/i.test(file)) {
    type = 'document';
    const title = base.replace(/\.(md|markdown)$/i, '');
    summary = `Dokumentasi: ${title} (${lines} baris).`;
    const d = DOC_DESCRIPTIONS[file];
    if (d) { summary = d.summary; tags = d.tags; }
  } else if (cat === 'infra' || /dockerfile|docker-compose|\.tf$/i.test(file)) {
    type = 'service';
  } else if (cat === 'script') {
    type = 'config';
  } else if (file === 'Makefile') {
    type = 'config';
  } else {
    type = 'config';
  }
  const d = DOC_DESCRIPTIONS[file];
  if (d) summary = d.summary;
  if (!summary) summary = `Berkas ${type} proyek: ${base}.`;
  if (!tags) tags = tagify(type, base.replace(/\.[^.]+$/, '').toLowerCase().replace(/[^a-z0-9-]+/g, '-'));
  const description = d && d.description ? d.description : summary;
  return { id: `${type}:${file}`, type, name: base, filePath: file, complexity: complexity(lines), summary, description, tags };
}

// Curated Indonesian descriptions for key non-code files
const DOC_DESCRIPTIONS = {
  'README.md': {
    summary: 'Dokumen utama proyek Polyglot — engine backend NetOps & manajemen ISP berbasis Go murni.',
    description: 'Menjelaskan platform: ConnectRPC, MCP, SSE/WebSocket, otomasi jaringan multi-vendor, billing ISP, hotspot voucher (parity Mikhmon), dan bot AI WhatsApp. Menjadi indeks seluruh dokumen teknis proyek.',
    tags: ['documentasi', 'overview', 'polyglot']
  },
  'AGENTS.md': {
    summary: 'Instruksi operasional AI agent: arsitektur repo, aturan non-negotiable, perintah build/test, dan definisi selesai.',
    description: 'Sumber kebenaran utama untuk konvensi: batasan layer (domain→port→usecase→adapter→driver), aturan error via pkg/response & fault.New, logger terpusat, batas 500 baris per berkas, dan checklist `make check`.',
    tags: ['documentasi', 'konvensi', 'agent', 'arsitektur']
  },
  'DEVELOPMENT-GUIDELINES.md': {
    summary: 'Panduan & standar pengembangan definitif: Clean Hexagonal, naming, interface, logging, error handling.',
    description: 'Rujukan standar implementasi harian: batasan arsitektur heksagonal, penamaan, desain interface Go, logging terpusat, dan penanganan error.',
    tags: ['documentasi', 'standar', 'arsitektur']
  },
  'EFFECTIVE_GO.md': {
    summary: 'Pedoman penulisan Go proyek: format, penamaan, error, dan desain paket.',
    description: 'Melengkapi EFFECTIVE GO resmi dengan konvensi spesifik repo Polyglot untuk penulisan kode Go yang konsisten.',
    tags: ['documentasi', 'go', 'konvensi']
  },
  'Polyglot-Architecture.md': {
    summary: 'Rasional arsitektur sistem: alur kerja, vendor-agnostic command gateway, dan state orchestration.',
    description: 'Dokumen arsitektur yang menjelaskan alur request antar-layer, konsep command gateway yang bebas vendor (MikroTik dll.), serta orkestrasi state perangkat.',
    tags: ['documentasi', 'arsitektur', 'gateway']
  },
  'SYSTEM-STRUCTURE-AND-ARCHITECTURE.md': {
    summary: 'Struktur folder definitif, relasi antar-komponen, dan diagram arsitektur sistem.',
    description: 'Peta struktur direktori internal/ (domain, port, usecase, adapter, driver), pkg/, api/, migrations/, beserta diagram relasi komponen.',
    tags: ['documentasi', 'struktur', 'arsitektur']
  },
  'PLAN-BACKEND-STANDARDIZATION-V2.md': {
    summary: 'Rencana pekerjaan standardisasi backend yang belum selesai.',
    description: 'Daftar sisa pekerjaan standardisasi backend: normalisasi layer, penyeragaman handler, dan penyelesaian migrasi arsitektur.',
    tags: ['documentasi', 'rencana', 'backend']
  },
  'docs/database-schema.md': {
    summary: 'Skema basis data manajemen ISP lengkap: pelanggan, subscription PPPoE/Hotspot, plans, invoices, transaksi.',
    description: 'Referensi skema database untuk domain ISP: pelanggan, langganan, paket layanan, tagihan, pembayaran, kas, dan laporan.',
    tags: ['documentasi', 'database', 'isp', 'billing']
  },
  'go.mod': {
    summary: 'Manifest modul Go (github.com/quixiq/polyglot, Go 1.26) dan seluruh dependensi.',
    description: 'Dependensi kunci: connectrpc.com/connect, modelcontextprotocol/go-sdk, gorm+casbin, redis, websocket (coder), whatsmeow, scrapligo, openai-go, firebase/genkit, golang-migrate, testcontainers.',
    tags: ['config', 'go', 'dependensi']
  },
  'Makefile': {
    summary: 'Target build & kualitas: build, vet, test, lint, check, proto-check, boundary checks, test-integration, security.',
    description: 'Gerbang kualitas proyek — make check menjalankan build, vet, test, lint, cek error connect, dan cek batasan layer; make proto-check memverifikasi kode generated protobuf.',
    tags: ['config', 'makefile', 'kualitas']
  },
  'api/openapi.yaml': {
    summary: 'Spesifikasi OpenAPI untuk permukaan HTTP API proyek.',
    description: 'Kontrak API HTTP yang di-serve di samping ConnectRPC — dipakai sebagai referensi endpoint REST dan dokumentasi.',
    tags: ['config', 'openapi', 'api']
  },
  'buf.yaml': {
    summary: 'Konfigurasi Buf (lint & breaking check) untuk modul protobuf api/proto.',
    description: 'Aturan lint dan deteksi breaking change protobuf; wajib lolos sebelum regenerasi kode (make proto-check).',
    tags: ['config', 'buf', 'protobuf']
  },
  'buf.gen.yaml': {
    summary: 'Template generasi kode protobuf (ConnectRPC + Go types) via buf.',
    description: 'Mengatur plugin generasi ke api/gen/ — direktori generated yang tidak boleh diedit manual.',
    tags: ['config', 'buf', 'codegen']
  },
  'deployments/docker-compose.yml': {
    summary: 'Komposisi layanan pengembangan: PostgreSQL, Redis, dan layanan pendukung.',
    description: 'Dipakai untuk menjalankan dependensi lokal (make db-up) sebelum menjalankan server atau test integrasi.',
    tags: ['infra', 'docker', 'dev']
  },
  'deployments/docker-compose.prod.yml': {
    summary: 'Komposisi layanan produksi: container aplikasi, PostgreSQL, Redis, reverse proxy.',
    description: 'Konfigurasi deployment produksi untuk engine Polyglot beserta dependensi layanannya.',
    tags: ['infra', 'docker', 'produksi']
  },
  '.env.example': {
    summary: 'Contoh konfigurasi environment: DATABASE_URL, REDIS, JWT, vendor MikroTik, WhatsApp, AI.',
    description: 'Template variabel lingkungan yang diperlukan server; salin ke .env untuk pengembangan lokal.',
    tags: ['config', 'env', 'konfigurasi']
  },
  '.golangci.yml': {
    summary: 'Konfigurasi golangci-lint — linter agregat untuk gate kualitas Go.',
    description: 'Menentukan linter aktif dan pengecualian; dijalankan oleh make lint.',
    tags: ['config', 'lint', 'go']
  },
};

// Edges for non-code files based on convention
function buildNonCodeEdges(f, allFiles, importMap) {
  const file = f.path;
  const cat = f.fileCategory;
  const edges = [];
  const isDoc = cat === 'docs' || /\.md$/i.test(file);
  const isInfra = cat === 'infra';
  const isCfg = cat === 'config' || cat === 'script' || file === 'Makefile';

  // documents → documents (architecture docs reference each other) & documents the areas they cover
  if (isDoc) {
    // doc → package area it documents
    const areaMap = [
      ['Polyglot-Architecture.md', 'internal/'],
      ['SYSTEM-STRUCTURE-AND-ARCHITECTURE.md', 'internal/'],
      ['DEVELOPMENT-GUIDELINES.md', 'internal/'],
      ['EFFECTIVE_GO.md', 'pkg/'],
      ['PLAN-BACKEND-STANDARDIZATION-V2.md', 'internal/'],
      ['docs/database-schema.md', 'migrations/'],
      ['docs/mikhmon/README.md', 'internal/'],
    ];
    for (const [doc, area] of areaMap) {
      if (file === doc) {
        const targets = allFiles.filter(x => x.path.startsWith(area) && x.fileCategory === 'code');
        const uniqPkgs = [...new Set(targets.map(t => t.path.split('/').slice(0, 2).join('/')))];
        for (const pkg of uniqPkgs.slice(0, 4)) {
          const rep = targets.find(t => t.path.startsWith(pkg + '/'));
          if (rep) edges.push({ source: `document:${file}`, target: `file:${rep.path}`, type: 'documents', weight: 0.5 });
        }
        break;
      }
    }
  }

  // infra configures code entrypoints
  if (isInfra && /docker-compose/.test(file)) {
    const main = allFiles.find(x => x.path === 'cmd/server/main.go');
    if (main) edges.push({ source: `${/prod/.test(file) ? 'service' : 'service'}:${file}`, target: `file:cmd/server/main.go`, type: 'configures', weight: 0.6 });
  }
  // Makefile runs everything
  if (file === 'Makefile') {
    const entry = allFiles.find(x => x.path === 'cmd/server/main.go');
    if (entry) edges.push({ source: 'config:Makefile', target: 'file:cmd/server/main.go', type: 'configures', weight: 0.6 });
  }
  return edges;
}

// ---------------------------------------------------------------------------
// Main pass
// ---------------------------------------------------------------------------

const allBatchFiles = batches.flatMap(b => b.files);
const scanFilesByPath = Object.fromEntries(allBatchFiles.map(f => [f.path, f]));

let report = { batches: 0, nodes: 0, edges: 0, skippedExtract: 0 };

for (const b of batches) {
  const idx = b.batchIndex;
  if (ONLY_BATCHES.length && !ONLY_BATCHES.includes(idx)) continue;

  const batchNodes = [];
  const batchEdges = [];

  // Partition files needing tree-sitter extraction (go/proto/sql)
  const goFiles = b.files.filter(f => f.language === 'go' && f.fileCategory === 'code');
  const protoFiles = b.files.filter(f => f.path.endsWith('.proto'));
  const sqlFiles = b.files.filter(f => f.path.endsWith('.sql'));
  const rest = b.files.filter(f => !goFiles.includes(f) && !protoFiles.includes(f) && !sqlFiles.includes(f));

  // Run extractor for this batch when needed
  let resultsByPath = {};
  if (goFiles.length || protoFiles.length || sqlFiles.length) {
    const inputPath = path.join(TMP, `extract-input-${idx}.json`);
    const outputPath = path.join(TMP, `extract-output-${idx}.json`);
    fs.writeFileSync(inputPath, JSON.stringify({
      projectRoot: PROJECT_ROOT,
      batchFiles: b.files,
      batchImportData: b.batchImportData || {},
    }));
    try {
      execFileSync('node', [path.join(SKILL_DIR, 'extract-structure.mjs'), inputPath, outputPath], { stdio: ['ignore', 'ignore', 'pipe'], cwd: PROJECT_ROOT, timeout: 120000 });
      const out = JSON.parse(fs.readFileSync(outputPath, 'utf8'));
      for (const r of (out.results || [])) resultsByPath[r.path] = r;
    } catch (err) {
      console.error(`Warning: batch ${idx} extract-structure failed: ${String(err.stderr || err.message).split('\n')[0]}`);
      report.skippedExtract += goFiles.length + protoFiles.length + sqlFiles.length;
    }
  }

  // Build nodes/edges per category
  for (const f of b.files) {
    const res = resultsByPath[f.path];
    const imp = (importMap[f.path] || []);
    if (res && f.language === 'go') {
      const built = buildGoFile(res, imp);
      batchNodes.push(...built.nodes);
      batchEdges.push(...built.edges);
    } else if (res && f.path.endsWith('.proto')) {
      const built = buildProtoFile(res);
      batchNodes.push(...built.nodes);
      batchEdges.push(...built.edges);
    } else if (res && f.path.endsWith('.sql')) {
      const built = buildSqlFile(res);
      batchNodes.push(...built.nodes);
      batchEdges.push(...built.edges);
    } else {
      const node = buildNonCodeNode(f);
      batchNodes.push(node);
      batchEdges.push(...buildNonCodeEdges(f, b.files, importMap));
      // scripts/test files referencing code via import map (js tests in scripts/)
      for (const target of imp) {
        batchEdges.push({ source: node.id, target: `file:${target}`, type: node.type === 'document' ? 'documents' : 'depends_on', weight: 0.5 });
      }
    }
  }

  // _replaceTargetPath: turn {target: basename, _replaceTargetPath} into file:<path> node id
  for (const e of batchEdges) {
    if (e._replaceTargetPath) {
      e.target = `file:${e._replaceTargetPath}`;
      delete e._replaceTargetPath;
    }
  }

  fs.writeFileSync(path.join(INTER, `batch-${idx}.json`), JSON.stringify({ batchIndex: idx, nodes: batchNodes, edges: batchEdges }, null, 1));

  report.batches++;
  report.nodes += batchNodes.length;
  report.edges += batchEdges.length;
  console.error(`batch ${idx}: ${b.files.length} files → ${batchNodes.length} nodes, ${batchEdges.length} edges`);
}

console.error(`build-batch-graph: batches=${report.batches} nodes=${report.nodes} edges=${report.edges} skippedExtract=${report.skippedExtract}`);
console.log(JSON.stringify(report));
