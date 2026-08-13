// Determine the previous graph's function-filtering rule
const fs = require('fs');
const path = require('path');
const TMP = '/home/quixiq/projects/polyground/polyglot/.understand-anything/tmp';
const g = JSON.parse(fs.readFileSync('/home/quixiq/projects/polyground/polyglot/.understand-anything/knowledge-graph.json', 'utf8'));

// Build set of previous function node ids (by file|name)
const prevFns = new Set();
for (const n of g.nodes) {
  if (n.type === 'function') {
    const m = n.id.match(/^function:(.*):([^:]+)$/);
    if (m) prevFns.add(`${m[1]}|${m[2]}`);
  }
}

// Build set of internal callees from callGraphs
const internalCallees = new Set();
for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    for (const c of r.callGraph || []) internalCallees.add(`${r.path}|${c.callee}`);
  }
}

let total = 0, exported = 0, lower = 0, inPrev = 0, exportedInPrev = 0, calledInPrev = 0, lowerInPrev = 0;
const examples = { inGraphButNotExportedNotCalled: [], notInGraph: [] };

for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    for (const fn of r.functions || []) {
      total++;
      const key = `${r.path}|${fn.name}`;
      const isExported = /^[A-Z]/.test(fn.name);
      const isCalled = internalCallees.has(key);
      if (isExported) exported++;
      else lower++;
      if (prevFns.has(key)) {
        inPrev++;
        if (isExported) exportedInPrev++;
        if (isCalled) calledInPrev++;
        else lowerInPrev++;
        if (!isExported && !isCalled && examples.inGraphButNotExportedNotCalled.length < 5) {
          examples.inGraphButNotExportedNotCalled.push(key);
        }
      } else {
        if (examples.notInGraph.length < 8) examples.notInGraph.push(`${key} (exported=${isExported}, called=${isCalled})`);
      }
    }
  }
}
console.log('total extracted:', total);
console.log('exported:', exported, 'lowercase:', lower);
console.log('in previous graph:', inPrev, '= exportedInPrev:', exportedInPrev, '+ calledInPrev:', calledInPrev, '+ other:', inPrev - exportedInPrev - calledInPrev);
console.log('examples in graph but neither exported nor called:', JSON.stringify(examples.inGraphButNotExportedNotCalled, null, 1));
console.log('examples NOT in graph:', JSON.stringify(examples.notInGraph, null, 1));
