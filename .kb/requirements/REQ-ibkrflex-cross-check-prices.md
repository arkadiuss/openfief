---
title: The IBKR Flex connector must deliver each open position's markPrice as a cross-check price of its instrument in the row's currency, dated by reportDate or else the statement's last day, from the SUMMARY row or else from LOT rows that agree, and never as a price; an unreadable, non-positive or disagreeing mark must become a skipped ingest problem and leave the position's assertion intact.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The IBKR Flex connector must deliver each open position's markPrice as a cross-check price of its instrument in the row's currency, dated by reportDate or else the statement's last day, from the SUMMARY row or else from LOT rows that agree, and never as a price; an unreadable, non-positive or disagreeing mark must become a skipped ingest problem and leave the position's assertion intact.
logic_claims:
  - CLAIM-85C522007389533A
semantic_clauses:
  - The IBKR Flex connector must deliver each open position's markPrice as a cross-check price of its instrument in the row's currency, dated by reportDate or else the statement's last day, from the SUMMARY row or else from LOT rows that agree, and never as a price; an unreadable, non-positive or disagreeing mark must become a skipped ingest problem and leave the position's assertion intact.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 238160d8346a0fa826ffaee2638b5b4c3247b02f87f4ada79a8eb3c81abffaf5
semantic_inventory:
  - claim_key: CLAIM-85C522007389533A
    claim_text: The IBKR Flex connector must deliver each open position's markPrice as a cross-check price of its instrument in the row's currency, dated by reportDate or else the statement's last day, from the SUMMARY row or else from LOT rows that agree, and never as a price; an unreadable, non-positive or disagreeing mark must become a skipped ingest problem and leave the position's assertion intact
    role: normative
    status: modeled
    span:
      start: 0
      end: 389
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:37.423Z'
id: REQ-ibkrflex-cross-check-prices
type: req
---
The IBKR Flex connector must deliver each open position's markPrice as a cross-check price of its instrument in the row's currency, dated by reportDate or else the statement's last day, from the SUMMARY row or else from LOT rows that agree, and never as a price; an unreadable, non-positive or disagreeing mark must become a skipped ingest problem and leave the position's assertion intact.

## Context

D-22 makes IBKR Flex prices a cross-check of the market-data connector's prices; the Flex statement's OpenPositions markPrice is IBKR's own end-of-day price of each held instrument, one per instrument and statement. Delivering it in prices would replace EODHD's close and revalue holdings, which D-22 doesn't allow. tradePrice is an execution price, not an end-of-day one, so it is not read. Fixtures are synthetic (AGENTS.md). Refines REQ-AUTO-7374F67D85269872. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22

> Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51
