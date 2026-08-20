import { readFileSync, writeFileSync } from 'node:fs';

const partial = JSON.parse(readFileSync('.understand-anything/intermediate/scan-partial.json', 'utf8'));
const importMapResult = JSON.parse(readFileSync('.understand-anything/tmp/import-map-result.json', 'utf8'));

const scanResult = {
  name: 'polyglot',
  description:
    'NetOps + ISP management backend — standalone Go service exposing MCP, REST, WebSocket/SSE, dan ConnectRPC untuk multi-vendor network automation (MikroTik RouterOS via API & SSH, Cisco, Huawei OLT, ZTE OLT, NETCONF, GenieACS, WhatsApp gateway) dengan clean architecture (domain/usecase/port/adapter/driver).',
  languages: ['go', 'yaml', 'markdown', 'sql', 'protobuf', 'dockerfile', 'makefile', 'json', 'txt', 'conf', 'mod', 'sum', 'jsonc'],
  frameworks: ['Gin', 'ConnectRPC', 'Casbin', 'GORM', 'MCP (Model Context Protocol)', 'scrapligo', 'whatsmeow', 'goros (RouterOS API)', 'Docker', 'Docker Compose', 'GitHub Actions'],
  files: partial.files,
  totalFiles: partial.totalFiles,
  filteredByIgnore: partial.filteredByIgnore,
  estimatedComplexity: partial.estimatedComplexity,
  importMap: importMapResult.importMap,
};

writeFileSync('.understand-anything/intermediate/scan-result.json', JSON.stringify(scanResult, null, 2));
console.log('WROTE scan-result.json — files:', scanResult.files.length, 'importMap entries:', Object.keys(scanResult.importMap).length);
