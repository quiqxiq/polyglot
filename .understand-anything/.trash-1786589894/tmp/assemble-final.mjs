import { readFileSync, writeFileSync } from 'node:fs';

const assembled = JSON.parse(readFileSync('.understand-anything/intermediate/assembled-graph.json', 'utf8'));
const layers = JSON.parse(readFileSync('.understand-anything/intermediate/layers.json', 'utf8'));
const tour = JSON.parse(readFileSync('.understand-anything/intermediate/tour.json', 'utf8'));
const scan = JSON.parse(readFileSync('.understand-anything/intermediate/scan-result.json', 'utf8'));

const graph = {
  version: '1.0.0',
  project: {
    name: 'polyglot',
    languages: scan.languages,
    frameworks: scan.frameworks,
    description: scan.description,
    analyzedAt: new Date().toISOString(),
    gitCommitHash: '387723c2fea703150aba892adceed6e6820a4bfa',
  },
  nodes: assembled.nodes,
  edges: assembled.edges,
  layers,
  tour,
};

writeFileSync('.understand-anything/intermediate/assembled-graph.json', JSON.stringify(graph, null, 2));
console.log('ASSEMBLED FINAL:', graph.nodes.length, 'nodes,', graph.edges.length, 'edges,', graph.layers.length, 'layers,', graph.tour.length, 'tour steps');
