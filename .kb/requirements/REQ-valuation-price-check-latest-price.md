---
title: A cross-check price must be compared with the latest stored price of the same instrument on or before its date.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-111-1
  - lane:strict
text_ref: design-decisions:D-111#1
origin:
  kind: import
  ref: design-decisions:D-111#1
  recorded_at: '2026-10-29T10:53:17.000Z'
semantic_text: A cross-check price must be compared with the latest stored price of the same instrument on or before its date.
logic_claims:
  - CLAIM-56F478866717EAAB
semantic_clauses:
  - A cross-check price must be compared with the latest stored price of the same instrument on or before its date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7bba2efb33a67950c3501acbca6404fe4c443dc5bd2cfe88611abd859bf23672
semantic_inventory:
  - claim_key: CLAIM-56F478866717EAAB
    claim_text: A cross-check price must be compared with the latest stored price of the same instrument on or before its date
    role: normative
    status: modeled
    span:
      start: 0
      end: 110
id: REQ-valuation-price-check-latest-price
type: req
---
A cross-check price must be compared with the latest stored price of the same instrument on or before its date.

## Why this changed

The author's decision D-111 compares a cross-check price with the latest stored price of the same instrument on or before its date, and shows a pair in different currencies with its own status instead of skipping it (REQ-valuation-price-check-currency-mismatch). REQ-valuation-price-check compared a cross-check price only with a price of the same instrument, date and currency, so this requirement replaces it; the 1% flag rule is REQ-valuation-price-check-one-percent.

## Context

D-22 makes IBKR Flex prices a cross-check of the market-data prices. The author recorded D-109 to D-112 after the design session that answered the implementation's questions.

## Source

> A cross-check price is compared with the latest stored price of the same instrument on or before its date. When the two are in different currencies, the pair is shown with its own status instead of being skipped, because the market-data connector may be pricing another listing than the one held.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-111#1
