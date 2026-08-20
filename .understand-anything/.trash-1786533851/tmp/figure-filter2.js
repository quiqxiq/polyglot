// Characterize old graph function nodes: for each, is it exported? called? has params?
const fs = require('fs');
const path = require('path');
const TMP = '/home/quixiq/projects/polyground/polyglot/.understand-anything/tmp';
const g = JSON.parse(fs.readFileSync('/home/quixiq/projects/polyground/polyglot/.understand-anything/knowledge-graph.json', 'utf8'));

// Build extraction lookup: file|name -> {params, isExported}
const extraction = new Map();
const internalCallees = new Set();
for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    for (const fn of r.functions || []) {
      extraction.set(`${r.path}|${fn.name}`, fn);
    }
    for (const c of r.callGraph || []) internalCallees.add(`${r.path}|${c.callee}`);
  }
}

const fns = g.nodes.filter((n) => n.type === 'function');
let inExtraction = 0, notInExtraction = 0;
let exported = 0, lowercase = 0;
let withParams = 0, noParams = 0;
let called = 0, notCalled = 0;
const weird = [];

for (const n of fns) {
  const m = n.id.match(/^function:(.*):([^:]+)$/);
  if (!m) { weird.push('no-match: ' + n.id); continue; }
  const key = `${m[1]}|${m[2]}`;
  const ext = extraction.get(key);
  if (!ext) { notInExtraction++; weird.push('not-in-extraction: ' + n.id); continue; }
  inExtraction++;
  if (/^[A-Z]/.test(ext.name)) exported++; else { lowercase++; if (weird.length < 20) weird.push('lowercase-in-graph: ' + n.id); }
  if (ext.params && ext.params.length) withParams++; else noParams++;
  if (internalCallees.has(key)) called++; else notCalled++;
}

console.log('old graph function nodes:', fns.length);
console.log('in extraction:', inExtraction, 'not in extraction:', notInExtraction);
console.log('exported:', exported, 'lowercase:', lowercase);
console.log('with params:', withParams, 'no params:', noParams);
console.log('internally called:', called, 'not called:', notCalled);
console.log('--- interesting samples ---');
console.log(weird.slice(0, 25).join('\n'));
