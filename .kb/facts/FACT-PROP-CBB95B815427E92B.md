---
title: Trading Currency Pick Required = true
status: active
text_ref: docs/DESIGN.md#D-170
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-170
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_confirmation
property_key: trading_currency_pick_required
canonical_key: mcp-kibi-model-requirement:ingest.ticker_confirmation:trading_currency_pick_required:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5491D9D37B051211
claim_text: 'Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner''s pick of one of that exchange''s currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker''s entries keep waiting.'
origin:
  kind: agent
  recorded_at: '2026-11-01T23:38:14.023Z'
id: FACT-PROP-CBB95B815427E92B
type: fact
---
