---
title: The pre-commit hook must run the full check when a change touches an input every check reads, when a merge is being committed, or when the selector cannot place a changed path.
status: open
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
semantic_text: The pre-commit hook must run the full check when a change touches an input every check reads, when a merge is being committed, or when the selector cannot place a changed path.
logic_claims:
  - CLAIM-DD99E75C6415489D
semantic_clauses:
  - The pre-commit hook must run the full check when a change touches an input every check reads, when a merge is being committed, or when the selector cannot place a changed path.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 35d05a657ce94e25baed383589f7e055400d78c4dd91b58787149e0c8d9c7b3c
semantic_inventory:
  - claim_key: CLAIM-DD99E75C6415489D
    claim_text: The pre-commit hook must run the full check when a change touches an input every check reads, when a merge is being committed, or when the selector cannot place a changed path
    role: normative
    status: modeled
    span:
      start: 0
      end: 175
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:50.085Z'
id: REQ-tooling-hook-full-fallback
type: req
---
The pre-commit hook must run the full check when a change touches an input every check reads, when a merge is being committed, or when the selector cannot place a changed path.

## Context

Module files, the Taskfile, lint configuration, tools, scripts, contracts, generated code and embedded design assets can change the result of any check, and a path no rule names may be read by something the selector does not know: a check skipped wrongly costs more than one run needlessly, so the author asked for the full check when in doubt. Requested by the author through the orchestrating agent (modular, affected-only hook and CI); modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> The pre-commit hook checks only what a commit can affect: lint and tests of the changed Go packages and of every package that imports them, found from Go's import graph rather than from paths. A commit of docs or requirements alone runs no Go step, but still builds the docs and runs the staleness gate and kibi's staged check. A change every check reads (module files, lint or task configuration, tools, scripts, contracts, generated code, embedded assets) or one the selector can't place runs the full check. When in doubt, full.

docs/dev/sdlc.md#tooling
