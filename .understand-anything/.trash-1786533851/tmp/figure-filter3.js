// Per-file comparison: extraction function count vs old-graph function count
const fs = require('fs');
const path = require('path');
const TMP = '/home/quixiq/projects/polyground/polyglot/.understand-anything/tmp';
const g = JSON.parse(fs.readFileSync('/home/quixiq/projects/polyground/polyglot/.understand-anything/knowledge-graph.json', 'utf8'));

const graphByFile = {};
for (const n of g.nodes) {
  if (n.type === 'function') graphByFile[n.filePath] = (graphByFile[n.filePath] || 0) + 1;
}

const extByFile = {};
for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    extByFile[r.path] = (r.functions || []).length;
  }
}

// Files where extraction > 0 but graph = 0 (fully dropped) or partial
const allPaths = new Set([...Object.keys(graphByFile), ...Object.keys(extByFile)]);
let fullyDropped = 0, partial = 0, full = 0;
const droppedExamples = [];
for (const p of allPaths) {
  const e = extByFile[p] || 0;
  const gr = graphByFile[p] || 0;
  if (e === 0) continue;
  if (gr === 0) { fullyDropped++; if (droppedExamples.length < 15) droppedExamples.push(`${p} (extracted ${e})`); }
  else if (gr < e) { partial++; if (droppedExamples.length < 25) droppedExamples.push(`${p} (extracted ${e}, graph ${gr})`); }
  else full++;
}
console.log('files with extraction but fully dropped from graph:', fullyDropped);
console.log('files partially represented:', partial);
console.log('files fully represented:', full);
console.log('--- samples ---');
console.log(droppedExamples.join('\n'));
