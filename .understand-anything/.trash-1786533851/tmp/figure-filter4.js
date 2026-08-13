// For a few files, list extracted functions vs old-graph functions
const fs = require('fs');
const path = require('path');
const TMP = '/home/quixiq/projects/polyground/polyglot/.understand-anything/tmp';
const g = JSON.parse(fs.readFileSync('/home/quixiq/projects/polyground/polyglot/.understand-anything/knowledge-graph.json', 'utf8'));

const graphNames = new Map(); // filePath -> Set(name)
for (const n of g.nodes) {
  if (n.type === 'function') {
    if (!graphNames.has(n.filePath)) graphNames.set(n.filePath, new Set());
    graphNames.get(n.filePath).add(n.name);
  }
}

const targets = [
  'internal/adapter/redis/store.go',
  'api/gen/v1/auth.pb.go',
  'internal/driver/mikrotik/connect.go',
  'internal/config/config.go',
  'internal/app/app.go',
];

for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    if (!targets.includes(r.path)) continue;
    const inGraph = graphNames.get(r.path) || new Set();
    const ext = r.functions || [];
    const dropped = ext.filter((f) => !inGraph.has(f.name));
    console.log(`=== ${r.path} (${ext.length} extracted, ${inGraph.size} in graph)`);
    for (const f of ext) {
      const keep = inGraph.has(f.name);
      console.log(`  ${keep ? 'KEEP' : 'DROP'} ${f.name} (${f.startLine}-${f.endLine}, ${(f.params || []).length} params)`);
    }
    console.log('');
  }
}
