---
title: A change to a file that a docs page covers must also change a page covering that file in the same push, or carry a Docs-Unchanged trailer; tests and test data never count.
status: open
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
semantic_text: A change to a file that a docs page covers must also change a page covering that file in the same push, or carry a Docs-Unchanged trailer; tests and test data never count.
logic_claims:
  - CLAIM-EC277DB6E1B2445F
semantic_clauses:
  - A change to a file that a docs page covers must also change a page covering that file in the same push, or carry a Docs-Unchanged trailer; tests and test data never count.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 90ffe1975c1b08d225c468e4e14e80f029194627f19cc3dc49ceb346486562ec
semantic_inventory:
  - claim_key: CLAIM-EC277DB6E1B2445F
    claim_text: A change to a file that a docs page covers must also change a page covering that file in the same push, or carry a Docs-Unchanged trailer; tests and test data never count
    role: normative
    status: modeled
    span:
      start: 0
      end: 170
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: docs/dev/docs.md
  recorded_at: '2026-10-28T09:32:50.291Z'
rationale: The author chose a hard gate so that docs cannot drift from the code they explain; agents merge straight to main, so it runs in the commit-msg hook and in CI over every pushed range.
id: REQ-docs-stale-gate
type: req
---
A change to a file that a docs page covers must also change a page covering that file in the same push, or carry a Docs-Unchanged trailer; tests and test data never count.

## Context

The author chose a hard gate so that docs cannot drift from the code they explain; agents merge straight to main, so it runs in the commit-msg hook and in CI over every pushed range. Decided by the author when choosing how OpenFief documents itself (Hugo, GitHub Pages and a hard staleness gate).

## Source

> See the rules in docs/dev/docs.md.

docs/dev/docs.md
