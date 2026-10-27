---
title: The public API must be REST/JSON with an OpenAPI spec, versioned /v1, covering accounts, transactions, tax lots, holdings and valuations.
status: open
text_ref: design-decisions:D-6#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-6-2
  - lane:strict
semantic_text: The public API must be REST/JSON with an OpenAPI spec, versioned /v1, covering accounts, transactions, tax lots, holdings and valuations.
logic_claims:
  - CLAIM-334DB9F780A969A3
semantic_clauses:
  - The public API must be REST/JSON with an OpenAPI spec, versioned /v1, covering accounts, transactions, tax lots, holdings and valuations.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 19e100ebdcef138499c752efa398aa8146f45e26bd2a23ddcd7c26083d9fcc7c
semantic_inventory:
  - claim_key: CLAIM-334DB9F780A969A3
    claim_text: The public API must be REST/JSON with an OpenAPI spec, versioned /v1, covering accounts, transactions, tax lots, holdings and valuations
    role: normative
    status: modeled
    span:
      start: 0
      end: 136
origin:
  kind: import
  ref: design-decisions:D-6#2
  recorded_at: '2026-10-27T20:12:57.604Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-F7E15E59C985C20D
type: req
---
The public API must be REST/JSON with an OpenAPI spec, versioned /v1, covering accounts, transactions, tax lots, holdings and valuations.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it. MCP waits for M2's model router.

Source: OpenFief author decisions (DESIGN.md section 10) - D-6#2
