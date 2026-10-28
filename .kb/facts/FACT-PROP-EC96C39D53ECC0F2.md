---
title: Books Transactions By Ibkr Id = true
status: active
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ibkr_flex
property_key: books_transactions_by_ibkr_id
canonical_key: mcp-kibi-model-requirement:ingest.ibkr_flex:books_transactions_by_ibkr_id:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DEEB2ADF98D7E880
claim_text: In transactions mode, IBKR Flex stock and ETF executions must become buys and sells keyed by tradeID, currency-pair executions FX conversions, and deposits, withdrawals, other fees and dividends transactions keyed by transactionID; in opening mode those rows must be skipped.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:43.775Z'
id: FACT-PROP-EC96C39D53ECC0F2
type: fact
---
