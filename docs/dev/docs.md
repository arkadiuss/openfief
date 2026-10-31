---
title: Documentation
description: How the docs are organized, how they stay true to the code, and how they are built and published.
covers:
  - site/**
  - tools/docgen/**
  - scripts/docs-diagrams.mjs
---

Most of OpenFief is written by coding agents, and the author verifies their work by reading. A page that describes code which has since changed is worse than no page. So the docs are treated like code: what can be generated is generated on every build, and what has to be written by hand is tied to the code it explains, so neither can drift without a check failing.

The idea is "living documentation" (Cyrille Martraire's book of that name, 2019), on top of the "docs as code" practice of the Write the Docs community: docs live in the repository, change in the same commits as the code, and go through the same pipeline.

## Kinds of pages

The pages follow Diátaxis (Daniele Procida, grown out of the Divio documentation system around 2017), which splits documentation by what the reader needs. Each kind has one source of truth and one mechanism that keeps it true.

| Kind | What it answers | Where | Kept true by |
|---|---|---|---|
| Tutorial | What does it look like, end to end? | [quickstart.md](../quickstart.md), a tour of the public demo | Review against the demo |
| How-to guides | How do I do a task? | [guides/](../guides/) | The staleness gate below, through each guide's `covers` |
| Explanation | Why is it like this, and how does it work? | [concepts/](../concepts/), [architecture.md](../architecture.md), [pipeline.md](pipeline.md) | The staleness gate below |
| Reference | What exactly is there? | Generated at build time: code map, custom resources, tasks, [status](/status/), [requirements report](/requirements/) | Regenerated on every build; never committed |
| Decisions | What did the author decide? | [DESIGN.md](../DESIGN.md), [proposals/](../proposals/), [research.md](../research.md) | The design-docs skill and `task docs:check` |
| Rules | What must contributors and agents do? | [dev/](./) and [AGENTS.md](../../AGENTS.md) | Review |

```mermaid
flowchart LR
  subgraph written["Written by hand"]
    design["DESIGN.md<br/>decisions"]
    glossary["glossary.md<br/>terms"]
    concepts["concepts/, architecture.md<br/>explanation"]
  end
  subgraph repo["Repository facts"]
    code["Go packages"]
    crds["deploy/crds"]
    taskfile["Taskfile.yml"]
    kb["kibi .kb/<br/>requirements, tests"]
  end
  docgen["tools/docgen"]
  site["Docs site"]
  code --> docgen
  crds --> docgen
  taskfile --> docgen
  kb --> docgen
  docgen -- "reference pages,<br/>status, code map" --> site
  design --> site
  glossary --> site
  concepts --> site
  concepts -. "cites IDs" .-> design
  concepts -. "links terms" .-> glossary
  concepts -. "covers" .-> code
```

## How the site reads

The sidebar tells one story, in reading order, and the docs home lays out the same groups as cards. Both come from `params.nav` in [`site/hugo.toml`](../../site/hugo.toml), so they can't disagree:

1. **Get started:** the [introduction](../OVERVIEW.md), the [quickstart](../quickstart.md) and [how it works](../architecture.md). A newcomer reads these three in order.
2. **Guides:** tasks, step by step, such as [exposing OpenFief](../guides/expose.md) and [building a plugin](../guides/build-a-plugin.md).
3. **Concepts:** each area in depth, split into **Platform** (plugins, operator, egress, identity, read API, views) and **Money and records** (kernel, ledger, ingest, valuation, treasury bonds, portfolio).
4. **Reference**, **Design** and **Contributing**: generated pages, the decisions, and these rules.

Groups other than the first two fold away until the reader opens one or is on one of its pages. A page left out of `params.nav` appears under "More", so nothing is ever hidden; a path in `params.nav` that names no page fails the build. Status belongs in the reference and at the end of a page, never at the top: say what a thing does first, and mark what isn't built where it matters, in one plain sentence.

## One fact, one place

- **Decisions** live in DESIGN.md. Other pages cite their IDs, such as D-18, and never restate them.
- **Terms** live in the glossary. Other pages link to it on first use and never redefine a term.
- **Requirements and their evidence** live in kibi. The status page renders them; no page copies them.
- **Code facts** such as package lists, custom-resource fields and task names are generated, never typed.

## Writing a concept page

A concept page explains one area as built. It starts with front matter:

```yaml
---
title: Ledger
description: One plain sentence saying what the page explains.
covers:
  - internal/ledger/**
---
```

- `covers` lists the code the page explains, as paths with `*` (one path segment) and `**` (any depth). It is what the staleness gate checks.
- The body has no H1; the title is rendered as one. The sections are **(P)**: an introduction, then `Concepts`, `How it works`, `Decisions`, `Verification` and `Not built yet`.
- Concepts are explained for a newcomer, with where the idea comes from when that helps.
- Every concept page has at least one diagram, and the names in it match the code.
- Diagrams of parts and how they connect draw each part with the design system's [icons](design-system.md#icons): `core@{ icon: "fief:gate", form: "square", label: "Core", pos: "b", h: 40 }`. Every icon node keeps its text label. Diagrams of code (types, functions, states, sequences) stay plain.
- Link source files and other docs with relative links. The build resolves each one and fails on a broken one.
- Synthetic data only, as everywhere in the repository.

## The staleness gate

A commit that changes a file a page covers must also change a page that covers that file, in the same push, or say why not in a commit trailer:

```
Docs-Unchanged: renamed a private helper; behavior is the same
```

- Tests (`_test.go`) and test data (`testdata/`) never trigger the gate: they don't change behavior.
- Every Go package must be covered by at least one page, so new code can't arrive undocumented.
- The gate runs in the `commit-msg` hook on the staged changes, and in CI over every pushed range. CI checks pushes as well as pull requests because agents merge straight to `main`.

The gate compares paths, not meaning: it can't tell whether the new text is right. Review still decides that. What the gate guarantees is that nobody changes covered code without looking at the page that explains it.

```mermaid
flowchart TD
  start(["Commits in a push,<br/>or the staged change"]) --> files["Files each commit changes,<br/>minus tests and testdata"]
  files --> covered{"Covered by a page?"}
  covered -- "no" --> pass(["Passes"])
  covered -- "yes" --> touched{"A covering page changed<br/>in the same push?"}
  touched -- "yes" --> pass
  touched -- "no" --> trailer{"The commit says<br/>Docs-Unchanged: why"}
  trailer -- "yes" --> pass
  trailer -- "no" --> fail(["Fails, naming the files<br/>and the pages to update"])
```

## Build and publishing

`task docs:build` builds the site into `site/public`:

1. `tools/docgen` writes the reference pages from the Go packages, the custom-resource definitions, the Taskfile and kibi's coverage report. kibi also renders its requirements report.
2. Hugo's first pass collects every Mermaid block on every page, DESIGN.md and proposals included.
3. `scripts/docs-diagrams.mjs` renders each block to a light and a dark SVG in headless Chromium, in the design system's palette, with `design/icons` registered as the Mermaid icon pack `fief`. Each diagram is cached by a hash of its source.
4. Hugo's second pass builds the site with the SVGs inlined. The site runs no JavaScript and loads nothing from other hosts.

The build fails on a diagram that doesn't render, a relative link that resolves to nothing, or a link to a heading that doesn't exist. `task docs:serve` previews the site locally.

```mermaid
flowchart LR
  push(["Push to main"]) --> check["CI: task check<br/>includes docs:build"]
  check --> pages["Publish site/public<br/>to GitHub Pages"]
  check --> images["Images, demo deploy"]
  release(["Release"]) --> assets["release:assets<br/>openfief-docs-VERSION.tar.gz"]
  pr(["Pull request"]) --> prcheck["CI: task check<br/>docs built, not published"]
```

- Every pull request and every push builds the docs as part of `task check`, so a broken page fails CI like a failing test.
- Each push to `main` publishes the site to GitHub Pages from the same CI run that deploys the demo. It is served at [openfief.dev](https://openfief.dev/), the custom domain in the repository's Pages settings: the landing page at the root, the docs under [/docs/](https://openfief.dev/docs/). `docs.openfief.dev` has no DNS record: a redirect there needs a Cloudflare rule the DNS token can't create. The demo at `demo.openfief.dev` is separate.
- Each release attaches `openfief-docs-<version>.tar.gz`, the site as it was at that version, for self-hosters who run that release.
- Every link and asset is relative to its page (Hugo's `relativeURLs`). So one build works at the custom domain, at the GitHub Pages fallback address, and opened straight from disk out of the tarball. Docs pages link to generated pages by docs path, such as `/status/`, which the link hook resolves under `/docs/`.

The site is [Hugo](https://gohugo.io), pinned in the Taskfile like the other tools. Its small theme lives in `site/`: no theme dependency, no CDN, and a strict content security policy. Its look, Mermaid diagrams included, comes from the [design system](design-system.md): the tokens, self-hosted fonts, marks and mascot in `design/`, which Hugo mounts into the site without `design.go`, the Go package that embeds them for the built-in UI. `docs/` is mounted into the site unchanged under `/docs/`, so every page reads the same on GitHub and on the site. The landing page is `site/content/_index.md` with its own template, `site/layouts/home.html`. It reads as a story: what OpenFief is, how it works, how to build on it and where it's going, with where it stands last. Its decision counts there come from the generated status page, so they are as current as the build, and the build fails if it can't read them.

## Verifying the implementation

The docs double as the author's verification tool:

- **[Status](/status/)** lists every decision with the kibi requirements derived from it, how many implementing symbols and tests each has, and its evidence level. A decision with no implemented requirement is visibly not built.
- **[Requirements report](/requirements/)** is kibi's stricter view: a requirement counts as tested only with a fresh, passing receipt from `kibi prove` for an end-to-end test of its scenario, and unit or integration tests never count. The site built from `main` carries the receipts of that commit's end-to-end runs on k3d and kind, ingested by the docs build ([pipeline.md](pipeline.md#ci-on-pull-requests-and-main)); a local `task docs:build` has none unless `task e2e:prove` ran in the checkout. Its *below evidence floor* count says whether any requirement lacks the evidence its firmness asks for.
- **[Code map](/reference/code-map/)** draws the as-built dependencies between bounded contexts from the Go imports. Compare it with the designed context map in [code-organization.md](code-organization.md).
- **Concept pages** end with `Verification` (which tests prove which behavior) and `Not built yet` (decided but missing).
