import { readFileSync, writeFileSync } from 'node:fs';

const assembled = JSON.parse(readFileSync('.understand-anything/intermediate/assembled-graph.json', 'utf8'));
const validIds = new Set(assembled.nodes.map(n => n.id));

const tour = JSON.parse(readFileSync('.understand-anything/intermediate/tour.json', 'utf8'));

// Re-resolve every nodeId in every step by path suffix, so prefix type doesn't matter
const idForPath = (p) => [...validIds].find(id => id.endsWith(`:${p}`)) || null;

for (const step of tour) {
  const resolved = step.nodeIds.map(id => {
    const colon = id.indexOf(':');
    const path = colon >= 0 ? id.slice(colon + 1) : id;
    return idForPath(path) || id;
  });
  step.nodeIds = [...new Set(resolved)].filter(id => validIds.has(id));
}

// Explicitly rebuild steps 18 & 19
tour.find(s => s.order === 18).nodeIds = [
  'schema:api/proto/v1/device.proto',
  'schema:api/proto/v1/hotspot.proto',
  'config:api/openapi.yaml',
  'schema:migrations/000001_create_devices_table.up.sql',
].filter(id => validIds.has(id));

tour.find(s => s.order === 19).nodeIds = [
  'resource:deployments/docker/Dockerfile',
  'resource:deployments/docker-compose.yml',
  'resource:.github/workflows/ci.yml',
  'resource:Makefile',
].filter(id => validIds.has(id));

writeFileSync('.understand-anything/intermediate/tour.json', JSON.stringify(tour, null, 2));
console.log('FIXED tour.json');
for (const s of tour) console.log(`  ${s.order}. ${s.title} (${s.nodeIds.length} nodes)`);
