---
title: 'The full kibi gate must give the same answer for the same authored knowledge base: a failure that names a kibi compile defect (an authored relationship missing from the compiled KB, or a duplicate entity row in the paginated projection) must recompile the store once from the same files and let the second check decide, and any other failure must fail at once.'
status: open
text_ref: docs/dev/pipeline.md#the-kibi-gate
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-the-kibi-gate
  - lane:strict
semantic_text: 'The full kibi gate must give the same answer for the same authored knowledge base: a failure that names a kibi compile defect (an authored relationship missing from the compiled KB, or a duplicate entity row in the paginated projection) must recompile the store once from the same files and let the second check decide, and any other failure must fail at once.'
logic_claims:
  - CLAIM-708AC38A1DCF5BF9
semantic_clauses:
  - 'The full kibi gate must give the same answer for the same authored knowledge base: a failure that names a kibi compile defect (an authored relationship missing from the compiled KB, or a duplicate entity row in the paginated projection) must recompile the store once from the same files and let the second check decide, and any other failure must fail at once.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a5ce5af2aff5aea995c30a7b516336617ff76ee01835b6f4b763eaab16760ce0
semantic_inventory:
  - claim_key: CLAIM-708AC38A1DCF5BF9
    claim_text: 'The full kibi gate must give the same answer for the same authored knowledge base: a failure that names a kibi compile defect (an authored relationship missing from the compiled KB, or a duplicate entity row in the paginated projection) must recompile the store once from the same files and let the second check decide, and any other failure must fail at once'
    role: normative
    status: modeled
    span:
      start: 0
      end: 359
origin:
  kind: agent
  recorded_at: '2026-10-30T04:09:54.404Z'
id: REQ-tooling-kibi-check-compile-defect
type: req
---
The full kibi gate must give the same answer for the same authored knowledge base: a failure that names a kibi compile defect (an authored relationship missing from the compiled KB, or a duplicate entity row in the paginated projection) must recompile the store once from the same files and let the second check decide, and any other failure must fail at once.

## Context

kibi 2.36 sometimes publishes a store that misses records of the authored files: 3 of 30 rebuild-and-check runs of an unchanged tree failed under load, rechecking the same store failed the same way, and the next compile passed. CI failed once on eaf5189 for the same reason. The author asked for a deterministic gate that is never weaker; an authored problem survives a recompile, so the second check still fails it. Requested by the author through the orchestrating agent; modeled by Claude Code, not yet reviewed.

## Source

> kibi nondeterminism: seen 3 times ... Never make the gate weaker.

Author request relayed by the orchestrating agent; docs/dev/pipeline.md#the-kibi-gate
