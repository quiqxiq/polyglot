import { readFileSync, writeFileSync } from 'node:fs';

// 1. Fix layers: docs predicate should only match .md, templates stay in core layer
const layers = JSON.parse(readFileSync('.understand-anything/intermediate/layers.json', 'utf8'));
for (const layer of layers) {
  if (layer.id === 'layer:docs') {
    layer.nodeIds = layer.nodeIds.filter(id => !id.includes('/internal/templates/'));
    layer.description = 'Project docs: architecture, ADRs, system structure, tech stack, MikroTik command reference, and agent instructions.';
  }
}
writeFileSync('.understand-anything/intermediate/layers.json', JSON.stringify(layers, null, 2));

// 2. Fix missing tags on nodes
const assembled = JSON.parse(readFileSync('.understand-anything/intermediate/assembled-graph.json', 'utf8'));
let fixed = 0;
for (const n of assembled.nodes) {
  if (!n.tags || !Array.isArray(n.tags) || n.tags.length === 0) {
    n.tags = ['untagged'];
    fixed++;
  }
  if (!n.name) n.name = n.id.split(':').pop();
  if (!n.summary || !n.summary.trim()) n.summary = 'No summary available';
}
console.log('fixed missing tags/fields on', fixed, 'nodes');
writeFileSync('.understand-anything/intermediate/assembled-graph.json', JSON.stringify(assembled, null, 2));

// 3. Re-validate
import { execSync } from 'node:child_process';
try {
  execSync('node .understand-anything/tmp/ua-inline-validate.cjs .understand-anything/intermediate/assembled-graph.json .understand-anything/intermediate/review.json', { stdio: 'inherit' });
} catch (e) {
  console.log('validate exit:', e.status);
}
const review = JSON.parse(readFileSync('.understand-anything/intermediate/review.json', 'utf8'));
console.log('FINAL issues:', review.issues.length, 'warnings:', review.warnings.length);
review.issues.slice(0, 10).forEach(i => console.log('  ISSUE: ' + i));
