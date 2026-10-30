---
title: 'The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go''s import graph.'
status: open
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
semantic_text: 'The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go''s import graph.'
logic_claims:
  - CLAIM-EAD1408AC02201D4
semantic_clauses:
  - 'The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go''s import graph.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 91e5959b13ce20cbf85ed7fa8af05cd5691f4b31962eab93ca4db9788d86ff57
semantic_inventory:
  - claim_key: CLAIM-EAD1408AC02201D4
    claim_text: 'The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go''s import graph'
    role: normative
    status: modeled
    span:
      start: 0
      end: 258
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:25.929Z'
id: REQ-tooling-hook-affected-packages
type: req
---
The pre-commit hook must lint, vulnerability-check and test only the Go packages a staged change affects: the changed packages and every package that imports them directly or not, plus the packages whose tests import one of them, found from Go's import graph.

## Context

The full check took about six minutes per commit, most of it in packages the commit cannot reach. The import graph comes from go list, so a package that imports a changed one is always checked, whatever its path. CI still runs the whole task check, so the hook is a fast first gate, not the only one. Requested by the author through the orchestrating agent (modular, affected-only hook and CI); modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> The pre-commit hook checks only what a commit can affect: lint and tests of the changed Go packages and of every package that imports them, found from Go's import graph rather than from paths. A commit of docs or requirements alone runs no Go step, but still builds the docs and runs the staleness gate and kibi's staged check. A change every check reads (module files, lint or task configuration, tools, scripts, contracts, generated code, embedded assets) or one the selector can't place runs the full check. When in doubt, full.

docs/dev/sdlc.md#tooling

> The selector is a tested Go tool, and every hook and CI run prints what it selected and why.

docs/dev/sdlc.md#tooling
