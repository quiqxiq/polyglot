#!/usr/bin/env node
// Assembles the final KnowledgeGraph object: project meta + nodes + edges + layers + tour.
import fs from 'fs';
import path from 'path';

const ROOT = '/home/quixiq/projects/polyground/polyglot';
const INTER = path.join(ROOT, '.understand-anything/intermediate');

const graph = JSON.parse(fs.readFileSync(path.join(INTER, 'assembled-graph.json'), 'utf8'));
const layers = JSON.parse(fs.readFileSync(path.join(INTER, 'layers.json'), 'utf8'));
const tour = JSON.parse(fs.readFileSync(path.join(INTER, 'tour.json'), 'utf8'));
const scan = JSON.parse(fs.readFileSync(path.join(INTER, 'scan-result.json'), 'utf8'));

// Drop dangling layer/tour refs
const nodeIds = new Set(graph.nodes.map((n) => n.id));
for (const l of layers) {
  l.nodeIds = (l.nodeIds || []).filter((id) => nodeIds.has(id));
}
for (const s of tour) {
  s.nodeIds = (s.nodeIds || []).filter((id) => nodeIds.has(id));
}

const finalGraph = {
  version: '1.0.0',
  project: {
    name: scan.name || 'polyglot',
    languages: scan.languages || [],
    frameworks: scan.frameworks || [],
    description: scan.description || '',
    analyzedAt: new Date().toISOString(),
    gitCommitHash: '2ea72d76e92b95c05b0d9782d4ed713bb93130c1',
  },
  nodes: graph.nodes,
  edges: graph.edges,
  layers,
  tour,
};

fs.writeFileSync(path.join(INTER, 'assembled-graph.json'), JSON.stringify(finalGraph, null, 2));
console.log('Assembled final graph:', finalGraph.nodes.length, 'nodes,', finalGraph.edges.length, 'edges,', finalGraph.layers.length, 'layers,', finalGraph.tour.length, 'tour steps.');
