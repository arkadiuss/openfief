---
title: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them.
status: open
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
semantic_text: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them.
logic_claims:
  - CLAIM-0D1270AC0DDD9ED6
semantic_clauses:
  - The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: af3aaa8d662ad3cbffb751a58b6356dab02502e106228c2b7dcd99c0c54a2d7e
semantic_inventory:
  - claim_key: CLAIM-0D1270AC0DDD9ED6
    claim_text: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them
    role: normative
    status: modeled
    span:
      start: 0
      end: 290
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:45.515Z'
id: REQ-ingest-cross-check-prices
type: req
---
The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them.

## Context

The connector contract gains cross_check_prices (additive, buf breaking passes) because a price in prices replaces the stored price, which is undecided shared market data (implementation questions); IBKR's marks must only be compared (D-22). The instance as source keeps two IBKR instances apart (D-48). Refines REQ-AUTO-7374F67D85269872 and REQ-valuation-cross-check-prices. Modeled by Claude Code for the D-22 cross-check task (impl/cross-check); the author has not reviewed it yet.

## Source

> Prices come from a market-data connector plugin where the user brings the API key (D-51); IBKR Flex prices serve as a cross-check.

docs/DESIGN.md D-22
