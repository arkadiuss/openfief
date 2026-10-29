---
title: A price check must compare a cross-check price only with the price of the same instrument, date and currency, must give the cross-check price's relative difference from that price rounded half-even to six decimal places, and must be flagged exactly when the absolute difference exceeds the threshold times the price.
status: closed
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: A price check must compare a cross-check price only with the price of the same instrument, date and currency, must give the cross-check price's relative difference from that price rounded half-even to six decimal places, and must be flagged exactly when the absolute difference exceeds the threshold times the price.
logic_claims:
  - CLAIM-03E78053777F2006
semantic_clauses:
  - A price check must compare a cross-check price only with the price of the same instrument, date and currency, must give the cross-check price's relative difference from that price rounded half-even to six decimal places, and must be flagged exactly when the absolute difference exceeds the threshold times the price.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c95eb91a8d260a175307753ccdc79c9cde73943fc96290ba475d7e090b5a84eb
semantic_inventory:
  - claim_key: CLAIM-03E78053777F2006
    claim_text: A price check must compare a cross-check price only with the price of the same instrument, date and currency, must give the cross-check price's relative difference from that price rounded half-even to six decimal places, and must be flagged exactly when the absolute difference exceeds the threshold times the price
    role: normative
    status: modeled
    span:
      start: 0
      end: 315
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:21.573Z'
id: REQ-valuation-price-check
type: req
---
A price check must compare a cross-check price only with the price of the same instrument, date and currency, must give the cross-check price's relative difference from that price rounded half-even to six decimal places, and must be flagged exactly when the absolute difference exceeds the threshold times the price.

## Context

D-22 says IBKR Flex prices serve as a cross-check of the market-data connector's prices, and D-51 adds that EODHD's closes are aggregated rather than each exchange's official close. Comparing only the same date and currency is the narrowest reading: another currency would need a conversion D-22 doesn't ask for. The flag is computed on the exact difference so that rounding for display never hides a discrepancy. No float64 for prices (AGENTS.md). Refines REQ-AUTO-7374F67D85269872. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22

> Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51
