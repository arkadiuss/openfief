---
title: The Ask box evaluation guide opens with a results table that docgen writes from the committed results, the bench's history and a one-line comments file, with one row per model and host for its latest full run that is not void, ordered passes first, then cases passed, then cost, and the docs build fails while the table in the guide is not what those files make.
status: open
firmness: default
required_evidence: unverified
tags:
  - askeval
origin:
  kind: agent
  ref: 'The author''s request on the models docs: a table with results, not prose only, eventually comments (2026-11-15)'
  recorded_at: '2026-11-15T19:31:40.431Z'
semantic_text: The Ask box evaluation guide opens with a results table that docgen writes from the committed results, the bench's history and a one-line comments file, with one row per model and host for its latest full run that is not void, ordered passes first, then cases passed, then cost, and the docs build fails while the table in the guide is not what those files make.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 443beb006a4c6717805aebc89f07350c73ae6f7b3c916209ce03e3b2283a16b4
semantic_inventory:
  - claim_key: CLAIM-91DE1A5E4B2BE33A
    claim_text: The Ask box evaluation guide opens with a results table that docgen writes from the committed results, the bench's history and a one-line comments file, with one row per model and host for its latest full run that is not void, ordered passes first, then cases passed, then cost, and the docs build fails while the table in the guide is not what those files make
    role: descriptive
    status: missing
    span:
      start: 0
      end: 361
logic_claims:
  - CLAIM-91DE1A5E4B2BE33A
id: REQ-askeval-guide-leads-with-generated-table
type: req
---
The Ask box evaluation guide opens with a results table that docgen writes from the committed results, the bench's history and a one-line comments file, with one row per model and host for its latest full run that is not void, ordered passes first, then cases passed, then cost, and the docs build fails while the table in the guide is not what those files make.

## Context

The guide was mostly prose. The columns are the model, host, verdict with its scorer version, cases passed, safety cases passed, reliability, USD per 1,000 questions, D-221's consecutive passes, the last run and a comment. A void run (no call answered, or from before the verdict) and a diagnostic run are not in the table; the results page keeps every run. The comments live in tools/askbench/comments.yaml, keyed by model@host, model@provider or the model alone.

## Source

> On models docs: I would like to rather see a table with results not prose only, eventually comments. (the author, 2026-11-15)