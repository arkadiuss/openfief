---
title: The built-in UI's tickers page must offer a choice of trading currency for a ticker whose exchange quotes in more than one, required when the proposal names no currency, and hand the owner's pick to the core with the ISIN.
status: open
text_ref: docs/DESIGN.md#D-170
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-170
  - lane:strict
semantic_text: The built-in UI's tickers page must offer a choice of trading currency for a ticker whose exchange quotes in more than one, required when the proposal names no currency, and hand the owner's pick to the core with the ISIN.
logic_claims:
  - CLAIM-1D055F1C9C41AAA6
semantic_clauses:
  - The built-in UI's tickers page must offer a choice of trading currency for a ticker whose exchange quotes in more than one, required when the proposal names no currency, and hand the owner's pick to the core with the ISIN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3a10a0fab6f96c4cc2063626f6a85e4c5215380eeb38cbc0631a5cb6fb886c45
semantic_inventory:
  - claim_key: CLAIM-1D055F1C9C41AAA6
    claim_text: The built-in UI's tickers page must offer a choice of trading currency for a ticker whose exchange quotes in more than one, required when the proposal names no currency, and hand the owner's pick to the core with the ISIN
    role: normative
    status: modeled
    span:
      start: 0
      end: 221
origin:
  kind: agent
  recorded_at: '2026-11-01T23:38:25.341Z'
id: REQ-ui-ticker-currency-choice
type: req
---
The built-in UI's tickers page must offer a choice of trading currency for a ticker whose exchange quotes in more than one, required when the proposal names no currency, and hand the owner's pick to the core with the ISIN.

## Context

D-170 has the owner pick the trading currency when confirming the ticker's ISIN in the UI (D-116). Recorded by the implementation agent of D-170 (impl/d170) with default firmness: no firm requirement for D-170 existed in kibi when it was built, so this traces the code to the decision directly until the design session imports the firm one.

## Source

> When the listing gives no currency and the exchange its market suffix names quotes in more than one, as London does in pence, pounds and dollars, the owner picks the trading currency when confirming the ticker's ISIN (D-116), and its positions wait until then; no currency is assumed.

docs/DESIGN.md#D-170
