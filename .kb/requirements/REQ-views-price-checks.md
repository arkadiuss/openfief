---
title: The read API must serve every price check as GET /v1/price-checks with its date, instrument, both prices with their sources, its relative difference as a decimal string and its status, filtered by status on request, together with the threshold the checks were flagged at, and must add only new methods and messages to the contract.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The read API must serve every price check as GET /v1/price-checks with its date, instrument, both prices with their sources, its relative difference as a decimal string and its status, filtered by status on request, together with the threshold the checks were flagged at, and must add only new methods and messages to the contract.
logic_claims:
  - CLAIM-6E8AD9D00747FAFD
semantic_clauses:
  - The read API must serve every price check as GET /v1/price-checks with its date, instrument, both prices with their sources, its relative difference as a decimal string and its status, filtered by status on request, together with the threshold the checks were flagged at, and must add only new methods and messages to the contract.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 75445fb75ac29e3018f22824ca4ab30720ae8298b0ec796e670eb18c287b560b
semantic_inventory:
  - claim_key: CLAIM-6E8AD9D00747FAFD
    claim_text: The read API must serve every price check as GET /v1/price-checks with its date, instrument, both prices with their sources, its relative difference as a decimal string and its status, filtered by status on request, together with the threshold the checks were flagged at, and must add only new methods and messages to the contract
    role: normative
    status: modeled
    span:
      start: 0
      end: 330
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:09.006Z'
id: REQ-views-price-checks
type: req
---
The read API must serve every price check as GET /v1/price-checks with its date, instrument, both prices with their sources, its relative difference as a decimal string and its status, filtered by status on request, together with the threshold the checks were flagged at, and must add only new methods and messages to the contract.

## Context

D-22's cross-check is only useful if the owner and the owner's scripts can see it, and D-6 makes the read API how every view, the built-in UI included, reads the record; D-50 defines it in protobuf with Vanguard REST. The contract only grows (buf breaking). Refines REQ-AUTO-7374F67D85269872. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1` ... Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6
