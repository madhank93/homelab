---
name: homelab-docs
description: "Homelab docs site conventions for docs/ (Astro Starlight) — picking the section and page type for new content, component page layout, decision pages, frontmatter and sidebar order, section overview cards, src: links to repo files, the software inventory as the only place versions appear, Starlight components (Steps, Tabs, asides) and verifying with just docs-check and npm run build. Use whenever a task adds, moves or edits a page under docs/src/content/docs/, documents a new workload or recipe, bumps a version that the inventory lists, or writes release notes. Do NOT use for Go doc comments or READMEs — use golang-documentation. Do NOT use for the workload code itself — use homelab-cdk8s."
---

# Homelab docs

Pages are Markdown or MDX under `docs/src/content/docs/`, and the URL follows the path. The rules (links, versions, writing style, asides) are on [`reference/development/docs-site.md`](../../../docs/src/content/docs/reference/development/docs-site.md). Read it before writing a page; this skill is the procedure around it.

## Where it goes

| Content | Page type | Location |
|---|---|---|
| One in-cluster app | Component page | `workloads/<domain>/<app>.md`, domain matching the Go package in `workloads/` |
| Talos, Cilium, Argo CD, cert-manager | Component page | `platform/` |
| Bifrost VPS, NetBird, Authentik, public exposure | Component page | `cloud/` |
| How the parts connect (traffic, secrets, GitOps) | Explanation, with a diagram | `concepts/` |
| First install, step by step | Tutorial | `start/deploy/` (numbered files) |
| A fix for a known failure | How-to | `operate/troubleshooting.md` (symptom heading) or `operate/repairs.md` (recipe row) |
| Moving a version | How-to | `operate/upgrades/` |
| Every recipe, every IP, every pin | Reference | `reference/justfile.md`, `reference/network.md`, `reference/software-inventory.md` |
| Why X over Y | Decision | `reference/decisions/<topic>.md` |
| What shipped | Release notes | `releases/v0-1-N.md(x)` |

One page, one type. A component page that grows a long "why" section gets a decision page and keeps one sentence plus a link.

## Add a component page

1. Copy the shape of `workloads/support/reloader.md`: frontmatter, one-sentence summary with an upstream link, at-a-glance table, then **How it fits** → **Configuration** (non-default settings only) → **Operate** → **Troubleshooting** (`###` per symptom). Drop empty sections.
2. Frontmatter: `title`, `description` (≤ 160 chars, one sentence), `sidebar.order`. Orders step by 10 so a page can slot in between.
3. Link the source as `` [`x.go`](src:workloads/<domain>/x.go) ``, never a GitHub URL. The file must be committed; nothing under `app/` is.
4. Add a `LinkCard` to the section's `index.mdx` overview.
5. A brand-new directory also needs a `group(...)` line in the sidebar in `docs/astro.config.mjs` and its own `index.mdx` (`sidebar.label: Overview`).
6. If it adds a pinned version: one inventory row (below), no version on the page.
7. If it adds a recipe or IP: a row in `reference/justfile.md` or `reference/network.md`.

## Add a decision page

Follow `reference/decisions/cnpg.md`: `**Decision:**` one sentence, `**Context:**` one paragraph, an options table (`Option | For | Against`, chosen option in bold), `**Consequences:**` bullets. No sidebar order; the group sorts alphabetically.

## Versions

Only `reference/software-inventory.md` states versions. A row is

```markdown
| Cilium | 1.18.12 | [`cilium.go`](src:core/platform/cilium.go) |
```

and the version must appear in that file (or in `config/versions.yaml` when the Go file reads `config.HelmChart` / `config.ContainerImage`). `same` in the last cell reuses the row above. On a bump, change the pin and the row together.

Everywhere else, describe behaviour and link `/reference/software-inventory/`. A version that names a past boundary ("added in v0.66") may stay on a line ending in `<!-- docs-check: historical -->`, or `{/* docs-check: historical */}` in MDX. `operate/upgrades/` and `releases/` are exempt.

## Components

- Plain `.md` unless the page imports a component; then rename to `.mdx` and import from `@astrojs/starlight/components`.
- `<Steps>` for procedures of three or more steps; `<Tabs>` only for parallel alternatives; `CardGrid`/`LinkCard` on overview pages.
- Asides: `:::danger` data loss or outage, `:::caution` surprising failure, `:::tip`, `:::note`.
- Shell blocks: `` ```bash frame="terminal" ``. Diagrams: `` ```mermaid ``, line breaks as `<br/>`, never `\n`.
- Never paste a secret value, token or key into an example; use `<placeholder>` and the `just openbao-*` recipes.

## Verify

```bash
just docs-check                # versions, manifests branch, src: links, mermaid
(cd docs && npm run build)     # broken internal links and anchors fail here
```

Both run in CI (`.github/workflows/docs.yml`); only release branches deploy. `just docs-serve` previews at http://localhost:4321.
