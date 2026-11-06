---
title: Corrects Confirmed Tickers = true
status: active
text_ref: docs/DESIGN.md#D-153
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-153
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_service
property_key: corrects_confirmed_tickers
canonical_key: mcp-kibi-model-requirement:ingest.ticker_service:corrects_confirmed_tickers:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-38F3CE700270FE8D
claim_text: The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND.
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:49.058Z'
id: FACT-PROP-A670ACC20AAA4C16
type: fact
---
