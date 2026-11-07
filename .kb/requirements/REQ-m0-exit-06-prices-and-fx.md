---
title: Every PLN figure must use NBP table A, end-of-day prices must come from the market-data plugin, and where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance must be flagged.
status: open
text_ref: design-decisions:D-200#6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-6
  - lane:strict
semantic_text: Every PLN figure must use NBP table A, end-of-day prices must come from the market-data plugin, and where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance must be flagged.
logic_claims:
  - CLAIM-FD0552E9ACD5B892
semantic_clauses:
  - Every PLN figure must use NBP table A, end-of-day prices must come from the market-data plugin, and where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance must be flagged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4c6a8d592ea4bb2965ca6eed35fbb3b1f6e40638a68c3d6a052759b6f84d5290
semantic_inventory:
  - claim_key: CLAIM-FD0552E9ACD5B892
    claim_text: Every PLN figure must use NBP table A, end-of-day prices must come from the market-data plugin, and where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance must be flagged
    role: normative
    status: modeled
    span:
      start: 0
      end: 224
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#6
  recorded_at: '2026-11-07T03:31:50.297Z'
id: REQ-m0-exit-06-prices-and-fx
type: req
---
Every PLN figure must use NBP table A, end-of-day prices must come from the market-data plugin, and where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance must be flagged.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-6.

## Source

> **Prices and FX.** Every PLN figure uses NBP table A. End-of-day prices come from the market-data plugin; where IBKR Flex reports a price for the same instrument and date, a difference beyond a configured tolerance is flagged.

design-decisions:D-200#6 (docs/DESIGN.md, D-200, criterion 6)
