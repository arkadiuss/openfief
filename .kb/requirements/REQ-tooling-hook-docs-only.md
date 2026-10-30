---
title: A commit that changes only documentation or requirements must run no Go step in the pre-commit hook, and must still build the docs, run the docs staleness gate and run kibi's staged check.
status: open
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
semantic_text: A commit that changes only documentation or requirements must run no Go step in the pre-commit hook, and must still build the docs, run the docs staleness gate and run kibi's staged check.
logic_claims:
  - CLAIM-505A586F876E963C
semantic_clauses:
  - A commit that changes only documentation or requirements must run no Go step in the pre-commit hook, and must still build the docs, run the docs staleness gate and run kibi's staged check.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 790e7d44fb8052555f70ec41b9152937737ee5c2364843c2d18bf3da3f300d13
semantic_inventory:
  - claim_key: CLAIM-505A586F876E963C
    claim_text: A commit that changes only documentation or requirements must run no Go step in the pre-commit hook, and must still build the docs, run the docs staleness gate and run kibi's staged check
    role: normative
    status: modeled
    span:
      start: 0
      end: 187
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:37.853Z'
id: REQ-tooling-hook-docs-only
type: req
---
A commit that changes only documentation or requirements must run no Go step in the pre-commit hook, and must still build the docs, run the docs staleness gate and run kibi's staged check.

## Context

Docs and kibi commits are frequent and cannot change a Go package, image or manifest, but they can break the docs site, the staleness gate or traceability, so those three checks stay. Requested by the author through the orchestrating agent (modular, affected-only hook and CI); modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> The pre-commit hook checks only what a commit can affect: lint and tests of the changed Go packages and of every package that imports them, found from Go's import graph rather than from paths. A commit of docs or requirements alone runs no Go step, but still builds the docs and runs the staleness gate and kibi's staged check. A change every check reads (module files, lint or task configuration, tools, scripts, contracts, generated code, embedded assets) or one the selector can't place runs the full check. When in doubt, full.

docs/dev/sdlc.md#tooling
