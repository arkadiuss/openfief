---
title: No trading currency may be assumed for a ticker whose listing gives none on an exchange that quotes in more than one.
status: open
text_ref: design-decisions:D-170#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-170-3
  - lane:strict
semantic_text: No trading currency may be assumed for a ticker whose listing gives none on an exchange that quotes in more than one.
logic_claims:
  - CLAIM-2A3EAA627FF2D1E9
semantic_clauses:
  - No trading currency may be assumed for a ticker whose listing gives none on an exchange that quotes in more than one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 18fdacf2bd02f8fa54e340fc6eac47f3854a78e510e5219afdaad3dff7ecb346
semantic_inventory:
  - claim_key: CLAIM-2A3EAA627FF2D1E9
    claim_text: No trading currency may be assumed for a ticker whose listing gives none on an exchange that quotes in more than one
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 116
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-170#3
  recorded_at: '2026-11-02T02:36:07.793Z'
id: REQ-ingest-ambiguous-trading-currency-not-assumed
type: req
---
No trading currency may be assumed for a ticker whose listing gives none on an exchange that quotes in more than one.

## Why this changed

On 2026-11-01 the author recorded D-170, which narrows D-159. The decision gives the reason: an exchange such as London quotes in pence, pounds and dollars, so its market suffix doesn't settle the trading currency. REQ-ingest-ticker-trading-currency took the exchange's currency whenever the listing named none, pence for London, which D-170 no longer allows when the exchange quotes in more than one, so this requirement replaces it, and it is closed. The implementation agent of D-170 (impl/d170) recorded default requirements for it before a firm one existed: REQ-ingest-ticker-currency-picked, REQ-ingest-ticker-currency-pick-required and REQ-ui-ticker-currency-choice. The firm requirements for D-170 replace them, and they are closed; their scenarios and the code that implements them move here, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. The decision gives the reason: such an exchange's suffix doesn't settle the currency, as London quotes in pence, pounds and dollars. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the listing gives no currency and the exchange its market suffix names quotes in more than one, as London does in pence, pounds and dollars, the owner picks the trading currency when confirming the ticker's ISIN (D-116), and its positions wait until then; no currency is assumed.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-170#3
