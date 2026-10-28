---
title: The built-in UI's code must import no OpenFief package other than the shared kernel and the SDK, so it reads the record only through the public read API.
status: open
text_ref: docs/DESIGN.md#D-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-6
  - lane:strict
semantic_text: The built-in UI's code must import no OpenFief package other than the shared kernel and the SDK, so it reads the record only through the public read API.
logic_claims:
  - CLAIM-3F70E0E511494A22
semantic_clauses:
  - The built-in UI's code must import no OpenFief package other than the shared kernel and the SDK, so it reads the record only through the public read API.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f5496022d1a598436520b745a51f2ace74ca7cfe3862cf24d9d45ba856f643e1
semantic_inventory:
  - claim_key: CLAIM-3F70E0E511494A22
    claim_text: The built-in UI's code must import no OpenFief package other than the shared kernel and the SDK, so it reads the record only through the public read API
    role: normative
    status: modeled
    span:
      start: 0
      end: 152
origin:
  kind: agent
  recorded_at: '2026-10-28T19:50:07.794Z'
id: REQ-ui-read-boundary
type: req
---
The built-in UI's code must import no OpenFief package other than the shared kernel and the SDK, so it reads the record only through the public read API.

## Context

D-6: every view plugin, the built-in UI included, uses the public API.

## Source

> Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md#D-6
