---
title: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices, and only when the instance's manifest declares them, otherwise storing none and adding one skipped ingest problem; a delivery with declared ones must fail when no cross-check store is wired rather than drop them.
status: open
text_ref: docs/DESIGN.md#D-112
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-112
  - lane:strict
semantic_text: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices, and only when the instance's manifest declares them, otherwise storing none and adding one skipped ingest problem; a delivery with declared ones must fail when no cross-check store is wired rather than drop them.
logic_claims:
  - CLAIM-9BADB86B715A26EC
semantic_clauses:
  - The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices, and only when the instance's manifest declares them, otherwise storing none and adding one skipped ingest problem; a delivery with declared ones must fail when no cross-check store is wired rather than drop them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9d96f06ee42a11ecb0fbf68ae0ac3c7ccae400a851f92306eb7a4bad705c0413
semantic_inventory:
  - claim_key: CLAIM-9BADB86B715A26EC
    claim_text: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices, and only when the instance's manifest declares them, otherwise storing none and adding one skipped ingest problem; a delivery with declared ones must fail when no cross-check store is wired rather than drop them
    role: normative
    status: modeled
    span:
      start: 0
      end: 406
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:45.515Z'
id: REQ-ingest-cross-check-prices
type: req
---
The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices, and only when the instance's manifest declares them, otherwise storing none and adding one skipped ingest problem; a delivery with declared ones must fail when no cross-check store is wired rather than drop them.

## Context

D-22 makes IBKR's prices a cross-check; D-112 lets only a connector whose manifest declares it send them. Reporting undeclared ones as a problem instead of refusing the batch is the narrow reading.

## Source

> Only a connector whose manifest declares it may send cross-check prices. The owner can acknowledge a flagged check, which is recorded append-only, and the UI shows the latest check per instrument with its history.

docs/DESIGN.md#D-112
