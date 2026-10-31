---
title: Cross Contract Checked = true
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
property_key: cross_contract_checked
canonical_key: mcp-kibi-model-requirement:ingest.ticker_entries:cross_contract_checked:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1F6746954BA09933
claim_text: The connector contract must carry the tickers a delivery names without an ISIN and their transactions, and the core must refuse a ticker transaction that names no ticker of its delivery or books no valid transaction.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:29.375Z'
id: FACT-PROP-BD561EDB13CFB47B
type: fact
---
