import { writeFileSync } from 'node:fs';

const meta = {
  lastAnalyzedAt: new Date().toISOString(),
  gitCommitHash: '387723c2fea703150aba892adceed6e6820a4bfa',
  version: '1.0.0',
  analyzedFiles: 273,
};

writeFileSync('.understand-anything/meta.json', JSON.stringify(meta, null, 2));
console.log('WROTE meta.json');
console.log(JSON.stringify(meta, null, 2));
