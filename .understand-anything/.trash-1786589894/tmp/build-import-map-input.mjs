import { readFileSync, writeFileSync } from 'node:fs';

const partial = JSON.parse(readFileSync('.understand-anything/intermediate/scan-partial.json', 'utf8'));

const files = partial.files.map(f => ({
  path: f.path,
  language: f.language,
  fileCategory: f.fileCategory,
}));

const input = {
  projectRoot: '/home/quixiq/projects/polyground/polyglot',
  files,
};

writeFileSync('.understand-anything/tmp/import-map-input.json', JSON.stringify(input, null, 2));
console.log('WROTE import-map-input.json with', files.length, 'files');
