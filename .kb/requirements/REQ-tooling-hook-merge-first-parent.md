---
title: The pre-commit hook must check a merge being committed from its diff against its first parent with the same affected-check selection as any other commit, so a merge of a test-only branch runs that branch's checks and not the whole check.
status: open
text_ref: docs/dev/sdlc.md#gate-tiers
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-gate-tiers
  - lane:strict
semantic_text: The pre-commit hook must check a merge being committed from its diff against its first parent with the same affected-check selection as any other commit, so a merge of a test-only branch runs that branch's checks and not the whole check.
logic_claims:
  - CLAIM-D7A50B37DFE77C01
semantic_clauses:
  - The pre-commit hook must check a merge being committed from its diff against its first parent with the same affected-check selection as any other commit, so a merge of a test-only branch runs that branch's checks and not the whole check.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 18a978eda949feae64539f8734cb6abdef8ce0455e1254246a8b1c98356a5cb1
semantic_inventory:
  - claim_key: CLAIM-D7A50B37DFE77C01
    claim_text: The pre-commit hook must check a merge being committed from its diff against its first parent with the same affected-check selection as any other commit, so a merge of a test-only branch runs that branch's checks and not the whole check
    role: normative
    status: modeled
    span:
      start: 0
      end: 236
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:42.558Z'
id: REQ-tooling-hook-merge-first-parent
type: req
---
The pre-commit hook must check a merge being committed from its diff against its first parent with the same affected-check selection as any other commit, so a merge of a test-only branch runs that branch's checks and not the whole check.

## Context

The hook used to run the full check for every merge, about 37 minutes, which the fast-gate plan replaces: while a merge is committed the staged diff is exactly the merge's diff against HEAD, its first parent, so the same selection covers what the merge brings in. The author approved the fast-gate plan (the pre-commit hook as a fast tier, the full gate in CI, retried downloads); implemented by Claude Code on the build/fast-gate branch, and the author has not reviewed this wording yet. 

## Source

> A merge needs no special case: while it is being committed, the staged diff is the merge's diff against its first parent (`HEAD`), so the selection covers what the merge brings in.

docs/dev/sdlc.md#gate-tiers
