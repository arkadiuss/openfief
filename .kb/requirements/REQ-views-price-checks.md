---
title: The read API must serve every price check as GET /v1/price-checks with the cross-check price's date, the instrument, both prices with their sources, the price's own date, its relative difference as a decimal string, empty for a currency mismatch, its status, its threshold and whether the owner acknowledged it, filtered by status on request, together with the default threshold, and must add only new fields, values and methods to the contract.
status: open
text_ref: docs/DESIGN.md#D-111
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-111
  - lane:strict
semantic_text: The read API must serve every price check as GET /v1/price-checks with the cross-check price's date, the instrument, both prices with their sources, the price's own date, its relative difference as a decimal string, empty for a currency mismatch, its status, its threshold and whether the owner acknowledged it, filtered by status on request, together with the default threshold, and must add only new fields, values and methods to the contract.
logic_claims:
  - CLAIM-282A98BE4FFD3CB9
semantic_clauses:
  - The read API must serve every price check as GET /v1/price-checks with the cross-check price's date, the instrument, both prices with their sources, the price's own date, its relative difference as a decimal string, empty for a currency mismatch, its status, its threshold and whether the owner acknowledged it, filtered by status on request, together with the default threshold, and must add only new fields, values and methods to the contract.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bde4eba97e5e4802c2a9d8fd4300a64dffcc9fad9580b8296f4ecb3c78c11e0f
semantic_inventory:
  - claim_key: CLAIM-282A98BE4FFD3CB9
    claim_text: The read API must serve every price check as GET /v1/price-checks with the cross-check price's date, the instrument, both prices with their sources, the price's own date, its relative difference as a decimal string, empty for a currency mismatch, its status, its threshold and whether the owner acknowledged it, filtered by status on request, together with the default threshold, and must add only new fields, values and methods to the contract
    role: normative
    status: modeled
    span:
      start: 0
      end: 444
origin:
  kind: agent
  recorded_at: '2026-10-29T07:03:09.006Z'
id: REQ-views-price-checks
type: req
---
The read API must serve every price check as GET /v1/price-checks with the cross-check price's date, the instrument, both prices with their sources, the price's own date, its relative difference as a decimal string, empty for a currency mismatch, its status, its threshold and whether the owner acknowledged it, filtered by status on request, together with the default threshold, and must add only new fields, values and methods to the contract.

## Context

D-22 makes IBKR's prices a cross-check; D-110 adds per-instrument thresholds, D-111 compares with the latest price on or before the date and gives currency mismatches their own status, and D-112 adds acknowledgements. The read API carries all of it for the UI and scripts, additively.

## Source

> A cross-check price is compared with the latest stored price of the same instrument on or before its date. When the two are in different currencies, the pair is shown with its own status instead of being skipped, because the market-data connector may be pricing another listing than the one held.

docs/DESIGN.md#D-111
