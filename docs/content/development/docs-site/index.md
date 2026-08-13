+++
title = "This Documentation Site"
description = "How these docs are built, published, and kept honest about versions."
weight = 20
+++

## What builds it

[Zola](https://www.getzola.org/) with the [Goyo](https://github.com/hahwul/goyo)
theme, from `docs/` in the repo. The theme is a git submodule, so a fresh clone
needs `git submodule update --init --recursive`.

```bash
just docs-serve     # http://127.0.0.1:1111, live reload
just docs-check     # version drift guard, same one CI runs
```

Run the local Zola at the same version CI uses — it is set once as
`ZOLA_VERSION` in {{ src(path=".github/workflows/docs.yml") }} and mirrored by a
comment in `docs/config.toml`. This matters more than it sounds: 0.21 renamed
`[markdown] highlight_code` to the `[markdown.highlighting]` table and removed
several highlight themes, so a config that builds on one version can fail
outright on the other.

## How it publishes

A push touching `docs/**` on `main` or any `v*` release branch triggers **Deploy
Docs**, which runs the version check, builds, and pushes `docs/public` to the
`gh-pages` branch.

Both branch patterns publish to the same place with `keep_files: false`, so a
concurrency group serialises them — otherwise a slow run could overwrite a newer
deploy. Whichever finishes last still wins, so avoid pushing docs to `main` and
a release branch at the same time.

## Writing pages

**Link internally with `@/`**, not with a site-root path:

```markdown
[Cilium](@/platform/networking/index.md)
[the VRAM section](@/hardware/gpu/index.md#sharing-the-16-gb-of-vram)
```

Zola resolves `@/` at build time and **fails the build** on a broken page or
anchor. A `/platform/networking` style link is not checked at all and 404s
silently in production.

**Link to source with the `src` shortcode**, never a hand-written URL:

```markdown
{{/* src(path="core/platform/cilium.go") */}}
{{/* src(path="workloads/cdk8s.yaml", label="cdk8s.yaml") */}}
```

It builds the URL from `extra.source_ref` in `config.toml`, so cutting a release
is one line rather than an edit to every page.

**Do not write version numbers.** They belong in the
[Software Inventory](@/architecture/software-inventory.md) and nowhere else;
`just docs-check` fails the build otherwise. Describe what a component does and
link there for which release is running.

The exception is a version that is a historical fact rather than a claim about
what is deployed — say, the release a feature first appeared in, or the one that
made `/etc` read-only. Those cannot drift. Mark the line:

```markdown
Talos v1.10+ restricts `machine.files` writes. <!-- docs-check: historical -->
```

## After a release cut

Three things move together:

1. `extra.source_ref` and `edit_url` in `docs/config.toml`
2. The manifests branch in {{ src(path="core/platform/argocd.go") }} — the check
   compares it against every branch name the docs mention
3. Any version that actually changed, in the Software Inventory
