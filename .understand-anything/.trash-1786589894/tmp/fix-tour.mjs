import { readFileSync, writeFileSync } from 'node:fs';

const assembled = JSON.parse(readFileSync('.understand-anything/intermediate/assembled-graph.json', 'utf8'));
const validIds = new Set(assembled.nodes.map(n => n.id));

// Find actual IDs for given file paths regardless of prefix
const idForPath = (p) => [...validIds].find(id => id.endsWith(`:${p}`)) || null;

const tour = JSON.parse(readFileSync('.understand-anything/intermediate/tour.json', 'utf8'));

for (const step of tour) {
  // Re-resolve all nodeIds by path suffix for robustness
  const resolved = step.nodeIds
    .map(id => {
      // id looks like prefix:path — re-resolve
      const colon = id.indexOf(':');
      const path = id.slice(colon + 1);
      return idForPath(path) || id;
    })
    .filter(Boolean);
  step.nodeIds = [...new Set(resolved)].filter(id => validIds.has(id));
}

writeFileSync('.understand-anything/intermediate/tour.json', JSON.stringify(tour, null, 2));
console.log('FIXED tour.json');
for (const s of tour) console.log(`  ${s.order}. ${s.title} (${s.nodeIds.length} nodes)`);
