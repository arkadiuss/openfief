---
title: The pre-commit hook must run the full check when a change touches an input every check reads or when the selector cannot place a changed path, and must plan a merge being committed from its diff against its first parent like any other commit.
status: open
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
semantic_text: The pre-commit hook must run the full check when a change touches an input every check reads or when the selector cannot place a changed path, and must plan a merge being committed from its diff against its first parent like any other commit.
logic_claims:
  - CLAIM-5617E5D04B6C944C
semantic_clauses:
  - The pre-commit hook must run the full check when a change touches an input every check reads or when the selector cannot place a changed path, and must plan a merge being committed from its diff against its first parent like any other commit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8cc156a6980148bc43a2489ce3f7a6b820272e392d99e762313a9a475068e140
semantic_inventory:
  - claim_key: CLAIM-5617E5D04B6C944C
    claim_text: The pre-commit hook must run the full check when a change touches an input every check reads or when the selector cannot place a changed path, and must plan a merge being committed from its diff against its first parent like any other commit
    role: normative
    status: modeled
    span:
      start: 0
      end: 241
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:50.085Z'
id: REQ-tooling-hook-full-fallback
type: req
---
The pre-commit hook must run the full check when a change touches an input every check reads or when the selector cannot place a changed path, and must plan a merge being committed from its diff against its first parent like any other commit.

## Context

Module files, the Taskfile, lint configuration, tools, scripts, contracts, generated code and embedded design assets can change the result of any check, and a path no rule names may be read by something the selector does not know: a check skipped wrongly costs more than one run needlessly, so the full check runs when in doubt. A merge is no longer one of those cases: while it is committed the staged diff is its diff against the first parent, which the selector plans like any commit. Refined for the fast-gate plan; modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> A change every check reads (module files, lint or task configuration, tools, scripts, contracts, generated code, embedded assets) or one the selector can't place runs the full check. When in doubt, full. A merge is checked like any commit, from its diff against its first parent.

docs/dev/sdlc.md#tooling
