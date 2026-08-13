// Test candidate filter rules against the old graph's function node set
const fs = require('fs');
const path = require('path');
const TMP = '/home/quixiq/projects/polyground/polyglot/.understand-anything/tmp';
const g = JSON.parse(fs.readFileSync('/home/quixiq/projects/polyground/polyglot/.understand-anything/knowledge-graph.json', 'utf8'));

const oldGraphKeys = new Set();
for (const n of g.nodes) {
  if (n.type === 'function') {
    const m = n.id.match(/^function:(.*):([^:]+)$/);
    if (m) oldGraphKeys.add(`${m[1]}|${m[2]}`);
  }
}

const extraction = [];
for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    for (const fn of r.functions || []) {
      extraction.push({ key: `${r.path}|${fn.name}`, name: fn.name, span: (fn.endLine || 0) - (fn.startLine || 0) });
    }
  }
}

// Candidate rules: keep if exported OR span >= N
for (const N of [8, 9, 10, 11, 12, 14, 16]) {
  let match = 0, total = 0;
  const falsePos = [], falseNeg = [];
  for (const f of extraction) {
    total++;
    const keep = /^[A-Z]/.test(f.name) || f.span >= N;
    const inOld = oldGraphKeys.has(f.key);
    if (keep === inOld) match++;
    else if (keep && !inOld) { if (falsePos.length < 4) falsePos.push(f.key + ` span=${f.span}`); }
    else if (!keep && inOld) { if (falseNeg.length < 4) falseNeg.push(f.key + ` span=${f.span}`); }
  }
  console.log(`N=${N}: match ${match}/${total} (${((match/total)*100).toFixed(1)}%)`);
  console.log(`   falsePos(examples): ${falsePos.join(' | ')}`);
  console.log(`   falseNeg(examples): ${falseNeg.join(' | ')}`);
}

// Also test: exported OR span>=10, but also drop name==='init'
for (const [label, rule] of [
  ['exported OR span>=10, drop init', (f) => (/^[A-Z]/.test(f.name) || f.span >= 10) && f.name !== 'init'],
  ['exported OR span>=10, drop init/rawDescGZIP', (f) => (/^[A-Z]/.test(f.name) || f.span >= 10) && f.name !== 'init' && !/rawDescGZIP$/.test(f.name)],
]) {
  let match = 0;
  const falsePos = [], falseNeg = [];
  for (const f of extraction) {
    const keep = rule(f);
    const inOld = oldGraphKeys.has(f.key);
    if (keep === inOld) match++;
    else if (keep && !inOld) { if (falsePos.length < 3) falsePos.push(f.key + ` span=${f.span}`); }
    else if (!keep && inOld) { if (falseNeg.length < 3) falseNeg.push(f.key + ` span=${f.span}`); }
  }
  console.log(`${label}: match ${match}/${extraction.length}`);
  console.log(`   falsePos: ${falsePos.join(' | ')}`);
  console.log(`   falseNeg: ${falseNeg.join(' | ')}`);
}
