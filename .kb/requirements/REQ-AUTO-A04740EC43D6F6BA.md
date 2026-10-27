---
title: View plugins must read the record only through the public API, the built-in UI included.
status: open
text_ref: design-decisions:D-6#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-6-3
  - lane:strict
semantic_text: View plugins must read the record only through the public API, the built-in UI included.
logic_claims:
  - CLAIM-5D0F72E07FF82034
semantic_clauses:
  - View plugins must read the record only through the public API, the built-in UI included.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7ef55e1b7de9d2ca86add4d9da9310b22457683106b68d2fbe8ae7aa6a7f6c49
semantic_inventory:
  - claim_key: CLAIM-5D0F72E07FF82034
    claim_text: View plugins must read the record only through the public API, the built-in UI included
    role: normative
    status: modeled
    span:
      start: 0
      end: 87
origin:
  kind: import
  ref: design-decisions:D-6#3
  recorded_at: '2026-10-27T20:12:58.549Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-A04740EC43D6F6BA
type: req
---
View plugins must read the record only through the public API, the built-in UI included.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it. MCP waits for M2's model router.

Source: OpenFief author decisions (DESIGN.md section 10) - D-6#3
