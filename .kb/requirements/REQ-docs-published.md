---
title: Each push to main must publish the docs site to GitHub Pages, and each release must attach the docs site as of that version.
status: open
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
semantic_text: Each push to main must publish the docs site to GitHub Pages, and each release must attach the docs site as of that version.
logic_claims:
  - CLAIM-FC63ECE9E8176CA1
semantic_clauses:
  - Each push to main must publish the docs site to GitHub Pages, and each release must attach the docs site as of that version.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 26baa302c9554bc4e4a7095f94aaa41e0a8ae59a932581daac6119fa4ae089a1
semantic_inventory:
  - claim_key: CLAIM-FC63ECE9E8176CA1
    claim_text: Each push to main must publish the docs site to GitHub Pages, and each release must attach the docs site as of that version
    role: normative
    status: modeled
    span:
      start: 0
      end: 123
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: docs/dev/docs.md
  recorded_at: '2026-10-28T09:33:15.138Z'
rationale: The author chose GitHub Pages from the same CI run that deploys the demo, plus a release asset so self-hosters can read the docs for the version they run.
id: REQ-docs-published
type: req
---
Each push to main must publish the docs site to GitHub Pages, and each release must attach the docs site as of that version.

## Context

The author chose GitHub Pages from the same CI run that deploys the demo, plus a release asset so self-hosters can read the docs for the version they run. Decided by the author when choosing how OpenFief documents itself (Hugo, GitHub Pages and a hard staleness gate).

## Source

> See the rules in docs/dev/docs.md.

docs/dev/docs.md
