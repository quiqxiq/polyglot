import { readFileSync, writeFileSync, readdirSync, statSync, existsSync } from 'node:fs';
import { join } from 'node:path';

const ROOT = '/home/quixiq/projects/polyground/polyglot';
const INT = join(ROOT, '.understand-anything/intermediate');
const TMP = join(ROOT, '.understand-anything/tmp');

const batches = JSON.parse(readFileSync(join(INT, 'batches.json'), 'utf8'));
const scanResult = JSON.parse(readFileSync(join(INT, 'scan-result.json'), 'utf8'));

// importMap: file path -> list of resolved internal paths (may be [] or [path])
const importMap = scanResult.importMap || {};
const allFilePaths = new Set(scanResult.files.map(f => f.path));

// Collect every known function name -> set of (path) that defines it, from extract results
const functionDefs = new Map(); // "path::name" -> {path, name}
const classDefs = new Map();

for (let i = 1; i <= 16; i++) {
  const f = join(TMP, `ua-file-extract-results-${i}.json`);
  if (!existsSync(f)) continue;
  const d = JSON.parse(readFileSync(f, 'utf8'));
  for (const r of d.results || []) {
    for (const fn of r.functions || []) functionDefs.set(`${r.path}::${fn.name}`, { path: r.path, name: fn.name });
    for (const cls of r.classes || []) classDefs.set(`${r.path}::${cls.name}`, { path: r.path, name: cls.name });
  }
}

// Global function name -> paths, to allow cross-file calls (same-name resolution, Go method receivers excluded)
const nameToPaths = new Map();
for (const [k, v] of functionDefs) {
  const nm = v.name;
  if (!nameToPaths.has(nm)) nameToPaths.set(nm, []);
  nameToPaths.get(nm).push(v.path);
}

function complexityFor(sizeLines) {
  if (sizeLines <= 50) return 'simple';
  if (sizeLines <= 200) return 'moderate';
  return 'complex';
}

function readHeader(path) {
  // Extract leading comment block (Go doc comments, or any file header) up to ~12 lines
  const full = join(ROOT, path);
  try {
    if (statSync(full).size > 200 * 1024) return '';
    const content = readFileSync(full, 'utf8').split('\n');
    const lines = [];
    for (const ln of content.slice(0, 80)) {
      const t = ln.trim();
      if (t === '') { if (lines.length > 0) break; continue; }
      if (t.startsWith('//')) { lines.push(t.replace(/^\/\/\s?/, '')); if (lines.length >= 10) break; continue; }
      if (t.startsWith('/*')) {
        const inner = t.replace(/^\/\*\*?/, '').replace(/\*\/$/, '').trim();
        if (inner) lines.push(inner);
        continue;
      }
      if (t.startsWith('*')) { lines.push(t.replace(/^\*\s?/, '')); if (lines.length >= 10) break; continue; }
      if (t.startsWith('#') || t.startsWith('--') || t.startsWith('<!--')) { lines.push(t); if (lines.length >= 6) break; continue; }
      // Stop at first code line
      if (lines.length > 0) break;
      if (/^(package |module |import |require |type |func |const |var |from |using |FROM |apiVersion|kind:)/.test(t)) break;
      if (lines.length === 0 && /^(package |module |import )/.test(t)) break;
    }
    return lines.join(' ').slice(0, 220);
  } catch {
    return '';
  }
}

function tagsFor(path) {
  const tags = new Set();
  const segs = path.split('/');
  if (path.endsWith('_test.go') || /\.test\./.test(path) || /(_test|\.spec\.|Test)/.test(path)) tags.add('test');
  if (path.includes('internal/')) tags.add('internal');
  for (const s of segs) {
    if (['adapter', 'domain', 'usecase', 'port', 'driver', 'config', 'registry', 'audit', 'app', 'migrations', 'pkg', 'cmd', 'docs', 'deployments', 'scripts', 'api', 'test', 'internal'].includes(s)) tags.add(s);
  }
  if (/\.go$/.test(path)) tags.add('go');
  if (/\.proto$/.test(path)) tags.add('protobuf');
  if (/\.yaml$|\.yml$/.test(path)) tags.add('yaml');
  if (/\.sql$/.test(path)) tags.add('sql');
  if (/\.md$/.test(path)) tags.add('documentation');
  if (/Dockerfile|docker-compose|\.yml$/.test(path) && path.includes('deploy')) tags.add('infra');
  return [...tags];
}

function summaryFor(path, sizeLines) {
  const header = readHeader(path);
  const base = header || path.split('/').pop().replace(/\.[^.]+$/, '').replace(/[_-]/g, ' ');
  let s = base;
  if (header && header.length < 100) {
    s = header;
  } else if (!header) {
    s = `${path.split('/').pop()} (${sizeLines} lines)`;
  }
  // Trim and ensure it reads like a sentence fragment
  return s.replace(/\s+/g, ' ').trim().slice(0, 260);
}

// ---- Build nodes & edges per batch ----
const typeToPrefix = {
  file: 'file', function: 'function', class: 'class', module: 'module', concept: 'concept',
  config: 'config', document: 'document', service: 'service', table: 'table', endpoint: 'endpoint',
  pipeline: 'pipeline', schema: 'schema', resource: 'resource',
};

// Map path -> node type based on fileCategory/language
function nodeTypeFor(f) {
  if (f.fileCategory === 'docs') return 'document';
  if (f.fileCategory === 'config') return 'config';
  if (f.fileCategory === 'infra') return 'resource';
  if (f.fileCategory === 'data') {
    if (/\.sql$/.test(f.path)) return 'schema';
    if (/\.proto$/.test(f.path)) return 'schema';
    return 'file';
  }
  return 'file';
}

const EDGE_WEIGHTS = { contains: 1.0, inherits: 0.9, implements: 0.9, calls: 0.8, exports: 0.8, defines_schema: 0.8, imports: 0.7, deploys: 0.7, migrates: 0.7, depends_on: 0.6, configures: 0.6, triggers: 0.6, tested_by: 0.5, documents: 0.5, provisions: 0.5, serves: 0.5, routes: 0.5, related: 0.5, subscribes: 0.5, publishes: 0.5, reads_from: 0.5, writes_to: 0.5, transforms: 0.5, validates: 0.5 };

const writtenFiles = new Set();

for (const b of batches.batches) {
  const nodes = [];
  const edges = [];
  const idx = b.batchIndex;
  const extFile = join(TMP, `ua-file-extract-results-${idx}.json`);
  const extData = existsSync(extFile) ? JSON.parse(readFileSync(extFile, 'utf8')) : { results: [] };
  const resultsByPath = new Map((extData.results || []).map(r => [r.path, r]));

  for (const f of b.files) {
    const path = f.path;
    const ntype = nodeTypeFor(f);
    const prefix = typeToPrefix[ntype] || 'file';
    const nid = `${prefix}:${path}`;
    const summary = summaryFor(path, f.sizeLines);
    const tags = tagsFor(path);
    const complexity = complexityFor(f.sizeLines);
    const node = { id: nid, type: ntype, name: path.split('/').pop(), summary, tags, complexity, filePath: path };
    if (ntype === 'config' || ntype === 'document' || ntype === 'resource' || ntype === 'schema') {
      node.name = path.split('/').pop();
    }
    nodes.push(node);
    writtenFiles.add(path);

    // Function & class nodes from structure extraction
    const ext = resultsByPath.get(path);
    if (ext) {
      for (const fn of ext.functions || []) {
        const fid = `function:${path}:${fn.name}`;
        const fnSummary = `Function ${fn.name}${fn.params && fn.params.length ? `(${fn.params.join(', ')})` : ''} in ${path}`;
        nodes.push({
          id: fid, type: 'function', name: fn.name,
          summary: fnSummary, tags: [...tags, 'function'], complexity: 'moderate',
          filePath: path,
        });
        edges.push({ source: nid, target: fid, type: 'contains', weight: 1.0, direction: 'forward' });
      }
      for (const cls of ext.classes || []) {
        const cid = `class:${path}:${cls.name}`;
        nodes.push({
          id: cid, type: 'class', name: cls.name,
          summary: `Type/class ${cls.name} in ${path}`, tags: [...tags, 'class'], complexity: 'moderate',
          filePath: path,
        });
        edges.push({ source: nid, target: cid, type: 'contains', weight: 1.0, direction: 'forward' });
      }

      // calls edges within this file
      for (const c of ext.callGraph || []) {
        const callerKey = `${path}::${c.caller}`;
        const calleeKey = `${path}::${c.callee}`;
        if (functionDefs.has(callerKey) && functionDefs.has(calleeKey)) {
          edges.push({
            source: `function:${path}:${c.caller}`,
            target: `function:${path}:${c.callee}`,
            type: 'calls', weight: 0.8, direction: 'forward',
          });
        } else if (functionDefs.has(callerKey)) {
          // callee may be external (stdlib, third-party) — link to file level only if internal file found
          const paths2 = nameToPaths.get(c.callee) || [];
          const internal = paths2.find(p => p !== path && allFilePaths.has(p));
          if (internal) {
            edges.push({
              source: `function:${path}:${c.caller}`,
              target: `function:${internal}:${c.callee}`,
              type: 'calls', weight: 0.8, direction: 'forward',
            });
          }
        }
      }
    }

    // imports edges (intra-project only)
    const imports = importMap[path] || [];
    for (const target of imports) {
      if (!allFilePaths.has(target)) continue;
      const ttype = nodeTypeFor(scanResult.files.find(x => x.path === target) || {});
      const tprefix = typeToPrefix[ttype] || 'file';
      edges.push({ source: nid, target: `${tprefix}:${target}`, type: 'imports', weight: 0.7, direction: 'forward' });
    }
  }

  writeFileSync(join(INT, `batch-${idx}.json`), JSON.stringify({ nodes, edges }, null, 2));
  console.log(`batch-${idx}.json: ${nodes.length} nodes, ${edges.length} edges`);
}

// Verify coverage
const missing = [...allFilePaths].filter(p => !writtenFiles.has(p));
console.log('=== coverage ===');
console.log('total scan files:', allFilePaths.size, 'covered:', writtenFiles.size, 'missing:', missing.length);
if (missing.length) console.log('missing sample:', missing.slice(0, 5));
