---
title: Request Instruments = held_instruments, model_portfolio_targets
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-93-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-93#1
fact_kind: property_value
subject_key: ingest.market_data_sync
property_key: request_instruments
operator: eq
value_type: set
value_set:
  - held_instruments
  - model_portfolio_targets
canonical_key: docs-design-md:ingest.market_data_sync:request_instruments:eq:held_instruments_model_portfolio_targets
claim_key: CLAIM-D3BA01BB08BEB655
claim_text: Each sync request to a market-data connector must carry the instruments the owner holds and the model portfolio's targets.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:46:44.958Z'
id: FACT-PROP-5C5191E181DBAB9F
type: fact
---
