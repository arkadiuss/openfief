---
title: Listing Or Market Exchange Currency = true
status: active
text_ref: design-decisions:D-159#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-159-2
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.trading_currency
property_key: listing_or_market_exchange_currency
canonical_key: mcp-kibi-model-requirement:ingest.trading_currency:listing_or_market_exchange_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-93E9CC51A173B944
claim_text: An XTB position's trading currency must be the currency of the confirmed ticker's listing or, when the listing gives none, the currency of the exchange its market suffix names.
origin:
  kind: agent
  recorded_at: '2026-11-02T00:54:14.837Z'
id: FACT-PROP-C135350DD37FAECA
type: fact
---
