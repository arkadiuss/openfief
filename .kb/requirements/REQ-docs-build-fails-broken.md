---
title: The docs build must fail on a relative link that resolves to nothing and on a Mermaid diagram that does not render.
status: open
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
semantic_text: The docs build must fail on a relative link that resolves to nothing and on a Mermaid diagram that does not render.
logic_claims:
  - CLAIM-E62A75C55B8ABFA1
semantic_clauses:
  - The docs build must fail on a relative link that resolves to nothing and on a Mermaid diagram that does not render.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6c30814a042c8815d075b6ce1a8ec091e85fd17e955a6ae93c0beec224ba2b4d
semantic_inventory:
  - claim_key: CLAIM-E62A75C55B8ABFA1
    claim_text: The docs build must fail on a relative link that resolves to nothing and on a Mermaid diagram that does not render
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: docs/dev/docs.md
  recorded_at: '2026-10-28T09:33:10.279Z'
rationale: The site is how the design and the shipped code are read; a broken link or diagram must stop the build like a failing test.
id: REQ-docs-build-fails-broken
type: req
---
The docs build must fail on a relative link that resolves to nothing and on a Mermaid diagram that does not render.

## Context

The site is how the design and the shipped code are read; a broken link or diagram must stop the build like a failing test. Decided by the author when choosing how OpenFief documents itself (Hugo, GitHub Pages and a hard staleness gate).

## Source

> See the rules in docs/dev/docs.md.

docs/dev/docs.md
