// Read at build time so node facts on the site cannot drift from the file
// Pulumi provisions from.
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parse } from 'yaml';

export interface Node {
  name: string;
  ip: string;
  role: 'control' | 'worker';
  cores: number;
  memory: number; // MiB
  disk: number; // GiB
  pcie_ids?: string[];
}

// Astro builds from docs/, and bundling moves this module, so resolve from the cwd.
const config = parse(readFileSync(resolve('../config/config.yml'), 'utf8'));

export const cluster = config.cluster as {
  name: string;
  vip: string;
  gateway_ip: string;
  lb_pool: { start: string; stop: string };
  nodes: Node[];
};
export const nodes = cluster.nodes;

export const totals = {
  nodes: nodes.length,
  controlPlanes: nodes.filter((n) => n.role === 'control').length,
  workers: nodes.filter((n) => n.role === 'worker').length,
  cores: nodes.reduce((sum, n) => sum + n.cores, 0),
  memoryGiB: Math.round(nodes.reduce((sum, n) => sum + n.memory, 0) / 1024),
  gpus: nodes.filter((n) => n.pcie_ids?.length).length,
};
