---
title: Every Go package must be covered by at least one docs page.
status: open
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
semantic_text: Every Go package must be covered by at least one docs page.
logic_claims:
  - CLAIM-6783FD5DC06CDC71
semantic_clauses:
  - Every Go package must be covered by at least one docs page.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f89aef67c5f5ccac1b016d2585446a1caa577b3cebfef8155553182c6c974bea
semantic_inventory:
  - claim_key: CLAIM-6783FD5DC06CDC71
    claim_text: Every Go package must be covered by at least one docs page
    role: normative
    status: modeled
    span:
      start: 0
      end: 58
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: docs/dev/docs.md
  recorded_at: '2026-10-28T09:32:55.161Z'
rationale: So that new code cannot arrive without a page that explains it.
id: REQ-docs-packages-covered
type: req
---
Every Go package must be covered by at least one docs page.

## Context

So that new code cannot arrive without a page that explains it. Decided by the author when choosing how OpenFief documents itself (Hugo, GitHub Pages and a hard staleness gate).

## Source

> See the rules in docs/dev/docs.md.

docs/dev/docs.md
