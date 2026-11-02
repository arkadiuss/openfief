---
title: When a ticker's listing gives no currency and the exchange its market suffix names quotes in more than one, the owner must pick the trading currency when confirming the ticker's ISIN.
status: open
text_ref: design-decisions:D-170#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-170-1
  - lane:strict
semantic_text: When a ticker's listing gives no currency and the exchange its market suffix names quotes in more than one, the owner must pick the trading currency when confirming the ticker's ISIN.
logic_claims:
  - CLAIM-CD8BE36807B6E5B5
semantic_clauses:
  - When a ticker's listing gives no currency and the exchange its market suffix names quotes in more than one, the owner must pick the trading currency when confirming the ticker's ISIN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 588d8b751cae592bf0bcc28ce13038556764e9b5cd3553769ae8193b7bc0b951
semantic_inventory:
  - claim_key: CLAIM-CD8BE36807B6E5B5
    claim_text: When a ticker's listing gives no currency and the exchange its market suffix names quotes in more than one, the owner must pick the trading currency when confirming the ticker's ISIN
    role: condition
    status: modeled
    span:
      start: 0
      end: 182
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-170#1
  recorded_at: '2026-11-02T00:56:25.554Z'
id: REQ-ingest-owner-picks-ambiguous-trading-currency
type: req
---
When a ticker's listing gives no currency and the exchange its market suffix names quotes in more than one, the owner must pick the trading currency when confirming the ticker's ISIN.

## Why this changed

On 2026-11-01 the author recorded D-170, which narrows D-159. The decision gives the reason: an exchange such as London quotes in pence, pounds and dollars, so its market suffix doesn't settle the trading currency. REQ-ingest-ticker-trading-currency took the exchange's currency whenever the listing named none, pence for London, which D-170 no longer allows when the exchange quotes in more than one, so this requirement replaces it, and it is closed.

## Context

The author recorded it on 2026-11-01. The decision gives the reason: London, for one, quotes in pence, pounds and dollars. The owner confirms the ISIN once (D-116). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the listing gives no currency and the exchange its market suffix names quotes in more than one, as London does in pence, pounds and dollars, the owner picks the trading currency when confirming the ticker's ISIN (D-116), and its positions wait until then; no currency is assumed.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-170#1
