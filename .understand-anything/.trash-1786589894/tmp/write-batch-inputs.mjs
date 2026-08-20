import { readFileSync, writeFileSync } from 'node:fs';

const batches = JSON.parse(readFileSync('.understand-anything/intermediate/batches.json', 'utf8'));

for (const b of batches.batches) {
  const input = {
    projectRoot: '/home/quixiq/projects/polyground/polyglot',
    batchFiles: b.files,
    batchImportData: b.batchImportData || {},
  };
  writeFileSync(`.understand-anything/tmp/ua-file-analyzer-input-${b.batchIndex}.json`, JSON.stringify(input, null, 2));
}

console.log('WROTE', batches.batches.length, 'batch input files');
