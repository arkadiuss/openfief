---
title: The built-in UI's overview must show the portfolio's total in the reporting currency, each account's valuation status, the allocation by asset class, currency, institution or sleeve, and the value over time from go-live, and must say that the time-weighted return is not built rather than compute performance.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: The built-in UI's overview must show the portfolio's total in the reporting currency, each account's valuation status, the allocation by asset class, currency, institution or sleeve, and the value over time from go-live, and must say that the time-weighted return is not built rather than compute performance.
logic_claims:
  - CLAIM-4D4F9DD029B571E5
semantic_clauses:
  - The built-in UI's overview must show the portfolio's total in the reporting currency, each account's valuation status, the allocation by asset class, currency, institution or sleeve, and the value over time from go-live, and must say that the time-weighted return is not built rather than compute performance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4b30e21e5b9284ea2c320cea6adc6cc51c0a1cf5c0ef673b743a5bc51878f3b4
semantic_inventory:
  - claim_key: CLAIM-4D4F9DD029B571E5
    claim_text: The built-in UI's overview must show the portfolio's total in the reporting currency, each account's valuation status, the allocation by asset class, currency, institution or sleeve, and the value over time from go-live, and must say that the time-weighted return is not built rather than compute performance
    role: normative
    status: modeled
    span:
      start: 0
      end: 308
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:57.000Z'
id: REQ-ui-overview
type: req
---
The built-in UI's overview must show the portfolio's total in the reporting currency, each account's valuation status, the allocation by asset class, currency, institution or sleeve, and the value over time from go-live, and must say that the time-weighted return is not built rather than compute performance.

## Context

D-30 (two charts), D-36 (valuation status in the UI), D-62 (reporting currency), D-61 (TWR, not built: the read API has no endpoint for it, so the UI shows 'TWR: not built'), D-52 (model portfolio off and hidden).

## Source

> The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

docs/DESIGN.md#D-30
