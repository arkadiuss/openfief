---
title: Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once.
status: open
text_ref: docs/DESIGN.md#D-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-6
  - lane:strict
semantic_text: Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once.
logic_claims:
  - CLAIM-6F18288550BE7608
semantic_clauses:
  - Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3b48c40de213449bed4fbde43c515e492b0dd50c64a23d9a71e574b7e08e1b84
semantic_inventory:
  - claim_key: CLAIM-6F18288550BE7608
    claim_text: Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once
    role: normative
    status: modeled
    span:
      start: 0
      end: 195
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:39.285Z'
id: REQ-views-transaction-pages
type: req
---
Transaction listings must be paged in date then source then external ID order with an opaque cursor and at most 500 transactions per page, so a client reads every current transaction exactly once.

## Context

D-6 puts transactions in the public API; an account accumulates thousands, so the listing is paged. The cursor names the last transaction returned, not an offset, so pages stay stable while the order holds. Page size bounds are the agent's choice. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6
