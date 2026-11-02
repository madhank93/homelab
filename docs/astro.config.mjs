import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';
import starlightImageZoom from 'starlight-image-zoom';
import starlightLinksValidator from 'starlight-links-validator';
import starlightLlmsTxt from 'starlight-llms-txt';
import remarkSrcLinks from './src/plugins/remark-src-links.mjs';

// The git ref these docs describe. Bump on a release cut: it drives the edit
// link and every `src:` link.
const repo = 'https://github.com/madhank93/homelab';
const ref = 'v0.1.8';

const group = (label, directory, collapsed = false) => ({
  label,
  collapsed,
  items: [{ autogenerate: { directory } }],
});

// Old Zola URLs, so links already shared keep working.
const redirects = {
  '/getting-started/': '/start/',
  '/getting-started/deployment/': '/start/deploy/',
  '/architecture/': '/concepts/',
  '/architecture/network-flow/': '/concepts/traffic/',
  '/architecture/kubernetes-architecture/': '/concepts/cluster/',
  '/architecture/gitops-flow/': '/concepts/gitops/',
  '/architecture/secrets-flow/': '/concepts/secrets/',
  '/platform/secrets/': '/concepts/secrets/',
  '/architecture/software-inventory/': '/reference/software-inventory/',
  '/hardware/cluster-nodes/': '/hardware/nodes/',
  '/infrastructure/': '/cloud/',
  '/infrastructure/hetzner-bifrost/': '/cloud/bifrost/',
  '/infrastructure/netbird/': '/cloud/netbird/',
  '/infrastructure/authentik/': '/cloud/authentik/',
  '/infrastructure/access/': '/cloud/exposure/',
  '/platform/proxmox-talos/': '/platform/talos/',
  '/platform/networking/': '/platform/cilium/',
  '/workloads/ai/nvidia-gpu-operator/': '/workloads/ai/nvidia-device-plugin/',
  '/workloads/management/headlamp/netbird-expose/': '/workloads/management/headlamp/',
  '/troubleshooting/': '/operate/troubleshooting/',
  '/upgrade-guide/': '/operate/upgrades/',
  '/glossary/': '/reference/glossary/',
  '/development/': '/reference/development/',
  '/development/devcontainer/': '/reference/development/devcontainer/',
  '/development/docs-site/': '/reference/development/docs-site/',
};

export default defineConfig({
  site: 'https://homelab.madhan.app',
  redirects,
  markdown: {
    remarkPlugins: [[remarkSrcLinks, { repo, ref }]],
  },
  integrations: [
    // Must precede starlight so its code fences become diagrams, not code blocks.
    mermaid({ autoTheme: true }),
    starlight({
      title: 'Homelab',
      logo: { src: './src/assets/logo.svg' },
      favicon: '/favicon.svg',
      head: [
        { tag: 'meta', attrs: { property: 'og:image', content: 'https://homelab.madhan.app/og.png' } },
        { tag: 'meta', attrs: { property: 'og:image:alt', content: 'Homelab: Kubernetes on Talos Linux and Proxmox' } },
        { tag: 'meta', attrs: { name: 'twitter:image', content: 'https://homelab.madhan.app/og.png' } },
      ],
      description:
        'Documentation for a Kubernetes homelab on Talos Linux, provisioned by Pulumi on Proxmox',
      social: [{ icon: 'github', label: 'GitHub', href: repo }],
      editLink: { baseUrl: `${repo}/edit/${ref}/docs/` },
      lastUpdated: true,
      customCss: ['./src/styles/custom.css'],
      plugins: [starlightLinksValidator(), starlightImageZoom(), starlightLlmsTxt()],
      sidebar: [
        {
          label: 'Start',
          items: ['start', 'start/quickstart', group('Deploy from scratch', 'start/deploy', true)],
        },
        group('Concepts', 'concepts'),
        group('Hardware', 'hardware'),
        group('Cloud', 'cloud'),
        group('Platform', 'platform'),
        {
          label: 'Workloads',
          items: [
            'workloads',
            group('AI', 'workloads/ai', true),
            group('Monitoring', 'workloads/monitoring', true),
            group('Security', 'workloads/security', true),
            group('Databases', 'workloads/databases', true),
            group('Secrets', 'workloads/secrets', true),
            group('Storage', 'workloads/storage', true),
            group('Registry', 'workloads/registry', true),
            group('Networking', 'workloads/networking', true),
            group('Automation', 'workloads/automation', true),
            group('Management', 'workloads/management', true),
            group('Support', 'workloads/support', true),
            group('Gaming', 'workloads/gaming', true),
          ],
        },
        {
          label: 'Operate',
          items: [
            'operate',
            'operate/troubleshooting',
            'operate/repairs',
            group('Upgrades', 'operate/upgrades', true),
          ],
        },
        {
          label: 'Reference',
          items: [
            'reference',
            'reference/software-inventory',
            'reference/justfile',
            'reference/network',
            'reference/glossary',
            group('Decisions', 'reference/decisions', true),
            group('Development', 'reference/development', true),
          ],
        },
        group('Releases', 'releases', true),
      ],
    }),
  ],
});
