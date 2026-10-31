---
title: Wait For Confirmed Isin = true
status: active
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_entries
property_key: wait_for_confirmed_isin
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:wait_for_confirmed_isin:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CA0492CD5AB6399A
claim_text: A ticker entry must be booked with the ISIN the owner confirmed for its ticker, and wait as an ingest problem until the owner confirms one.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:20.998Z'
id: FACT-PROP-EB6755D28B8A41D9
type: fact
---
