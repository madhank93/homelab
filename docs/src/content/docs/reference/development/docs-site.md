---
title: Docs site
description: How these docs are built with Astro Starlight, linked, kept honest about versions, deployed and written.
sidebar:
  order: 2
---

[Astro Starlight](https://starlight.astro.build/) in `docs/`; pages are Markdown or MDX under `docs/src/content/docs/`, and the URL follows the path (`workloads/storage/longhorn.md` → `/workloads/storage/longhorn/`).

```bash frame="terminal"
just docs-serve    # npm run dev, http://localhost:4321, live reload
just docs-check    # version and link guard, same as CI
```

Run `npm ci` in `docs/` once first.

## Links

- **Internal:** absolute with a trailing slash, `/platform/cilium/`. [starlight-links-validator](https://github.com/HiDeoo/starlight-links-validator) fails the build on a broken page or anchor.
- **Source:** `` [`cilium.go`](src:core/platform/cilium.go) ``. A remark plugin ([`remark-src-links.mjs`](src:docs/src/plugins/remark-src-links.mjs)) rewrites `src:` to the file on GitHub at the `ref` constant in [`astro.config.mjs`](src:docs/astro.config.mjs). The path must be git-tracked; `app/` never is.

## Versions

Versions live only in the [Software inventory](/reference/software-inventory/). `just docs-check` ([`docs-version-check.sh`](src:scripts/docs-version-check.sh)) fails when:

- an inventory version no longer appears in the file its row names;
- any other page states a version (`operate/upgrades/` and `releases/` are exempt);
- a page names a manifests branch other than `manifests_branch` in `config/config.yml`;
- a `src:` link points at an untracked file;
- a Mermaid label uses `\n` instead of `<br/>`.

A version that is a historical fact, not a claim about what runs, is allowed on a line ending in `<!-- docs-check: historical -->` (`{/* docs-check: historical */}` in MDX).

## Deploy

[`.github/workflows/docs.yml`](src:.github/workflows/docs.yml) runs on pushes to `main` and `v*` branches that touch docs, config or code. Both run the check and build; **only release branches deploy** `docs/dist` to `gh-pages`, so the live site matches the `ref` in `astro.config.mjs`.

On a release cut, move together: `ref` in `astro.config.mjs`, `manifests_branch` in `config/config.yml`, and any inventory row that changed.

## Writing style

- Short: present tense, active voice; runbooks imperative ("Run", "Check"). No "we", no marketing, no emoji.
- Only homelab-specific facts: one sentence on what a component does here, then a link to upstream. No "What is X?" sections.
- Tables for facts, prose for reasons; a reason is one sentence. Sentence-case headings; description ≤ 160 chars.
- Component pages follow: summary → at-a-glance table (namespace, source, exposure, storage, auth) → how it fits → configuration (non-default settings only) → operate → troubleshooting (symptom headings). Drop empty sections.
- "Why X over Y" goes in a decision page under `reference/decisions/`; the component page keeps one sentence and a link.
- Asides: `danger` for data loss or outage, `caution` for surprising failure, `tip`, `note`. Use `<Steps>` for procedures of three or more steps, `<Tabs>` only for parallel alternatives. Rename to `.mdx` only when using a component.
