#!/usr/bin/env node
// Enrich file-level node summaries with their exported symbols and import counts
// for better readability in the dashboard.
import fs from 'fs';
import path from 'path';

const ROOT = '/home/quixiq/projects/polyground/polyglot';
const GRAPH_PATH = path.join(ROOT, '.understand-anything/intermediate/assembled-graph.json');
const TMP = path.join(ROOT, '.understand-anything/tmp');

const graph = JSON.parse(fs.readFileSync(GRAPH_PATH, 'utf8'));

// Load extraction to get exports per file
const extractionByFile = {};
for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) extractionByFile[r.path] = r;
}

const nodeById = new Map(graph.nodes.map((n) => [n.id, n]));

// Which file nodes have function/class children (to derive exports)
const exportsByFile = new Map();
for (const n of graph.nodes) {
  if (n.type === 'function' || n.type === 'class') {
    const m = n.id.match(/^(?:function|class):(.*):([^:]+)$/);
    if (!m) continue;
    const filePath = m[1];
    if (!exportsByFile.has(filePath)) exportsByFile.set(filePath, []);
    exportsByFile.get(filePath).push(n.name);
  }
}

for (const n of graph.nodes) {
  if (n.type === 'file') {
    const exports = exportsByFile.get(n.filePath) || [];
    const ext = extractionByFile[n.filePath];
    const importCount = ext?.metrics?.importCount || 0;
    const fnCount = ext?.functions?.length || 0;
    const exported = exports.filter((e) => /^[A-Z]/.test(e)).slice(0, 8);
    let summary = `${n.filePath} — Go source file`;
    if (fnCount) summary += ` with ${fnCount} functions`;
    if (importCount) summary += `, ${importCount} imports`;
    if (exported.length) summary += `. Exports: ${exported.join(', ')}`;
    n.summary = summary;
  }
}

fs.writeFileSync(GRAPH_PATH, JSON.stringify(graph, null, 2));
console.log('Enriched', graph.nodes.length, 'nodes.');
