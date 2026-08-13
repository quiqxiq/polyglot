// Inspect non-code extraction shapes for proto, SQL, Dockerfile, workflows
const fs = require('fs');
const path = require('path');
const TMP = '/home/quixiq/projects/polyground/polyglot/.understand-anything/tmp';

for (let i = 1; i <= 16; i++) {
  const p = path.join(TMP, `ua-file-extract-results-${i}.json`);
  if (!fs.existsSync(p)) continue;
  const d = JSON.parse(fs.readFileSync(p, 'utf8'));
  for (const r of d.results) {
    if (r.path.includes('.proto')) {
      console.log('=== PROTO:', r.path);
      console.log('  definitions[0]:', JSON.stringify((r.definitions || [])[0] || null).slice(0, 300));
      console.log('  endpoints[0]:', JSON.stringify((r.endpoints || [])[0] || null).slice(0, 300));
    }
    if (r.path.includes('migrations/') && r.path.endsWith('.sql')) {
      console.log('=== SQL:', r.path);
      const defs = r.definitions || [];
      console.log('  definitions count:', defs.length);
      console.log('  definitions[0]:', JSON.stringify(defs[0] || null).slice(0, 300));
    }
    if (r.path.includes('Dockerfile')) {
      console.log('=== DOCKER:', r.path);
      console.log('  services[0]:', JSON.stringify((r.services || [])[0]).slice(0, 300));
      console.log('  steps[0]:', JSON.stringify((r.steps || [])[0]).slice(0, 200));
    }
    if (r.path.includes('.github/workflows')) {
      console.log('=== WORKFLOW:', r.path);
      console.log('  sections[0]:', JSON.stringify((r.sections || [])[0]).slice(0, 200));
    }
  }
}
