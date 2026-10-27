---
title: The MCP server must not be exposed before M2.
status: open
text_ref: design-decisions:D-6#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-6-4
  - lane:strict
semantic_text: The MCP server must not be exposed before M2.
logic_claims:
  - CLAIM-2E2392B15C00CB84
semantic_clauses:
  - The MCP server must not be exposed before M2.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dc834992504f15b35c4da81a16f29dca530aa8cc8391f3c805a432e1b44af2ec
semantic_inventory:
  - claim_key: CLAIM-2E2392B15C00CB84
    claim_text: The MCP server must not be exposed before M2
    role: normative
    status: modeled
    span:
      start: 0
      end: 44
origin:
  kind: import
  ref: design-decisions:D-6#4
  recorded_at: '2026-10-27T20:12:59.490Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-B8D1BCB44E147E0C
type: req
---
The MCP server must not be exposed before M2.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it. MCP waits for M2's model router.

Source: OpenFief author decisions (DESIGN.md section 10) - D-6#4
