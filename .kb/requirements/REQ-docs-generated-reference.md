---
title: The code map, custom-resource, task and status pages must be generated from the repository on every docs build and never committed.
status: open
text_ref: docs/dev/docs.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-docs-md
  - lane:strict
semantic_text: The code map, custom-resource, task and status pages must be generated from the repository on every docs build and never committed.
logic_claims:
  - CLAIM-A41D121C2372B9E8
semantic_clauses:
  - The code map, custom-resource, task and status pages must be generated from the repository on every docs build and never committed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fb631adf631adbce4240ba4dfa8fb379cea1ff924425188dfed9f312da111520
semantic_inventory:
  - claim_key: CLAIM-A41D121C2372B9E8
    claim_text: The code map, custom-resource, task and status pages must be generated from the repository on every docs build and never committed
    role: normative
    status: modeled
    span:
      start: 0
      end: 130
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: docs/dev/docs.md
  recorded_at: '2026-10-28T09:33:05.224Z'
rationale: Reference that is generated cannot drift; the status and code map pages are also how the author verifies what is built.
id: REQ-docs-generated-reference
type: req
---
The code map, custom-resource, task and status pages must be generated from the repository on every docs build and never committed.

## Context

Reference that is generated cannot drift; the status and code map pages are also how the author verifies what is built. Decided by the author when choosing how OpenFief documents itself (Hugo, GitHub Pages and a hard staleness gate).

## Source

> See the rules in docs/dev/docs.md.

docs/dev/docs.md
