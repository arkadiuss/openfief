---
title: Owner Corrects Wrong Isin Superseding Transactions = true
status: active
text_ref: design-decisions:D-153#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-153-2
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ticker_isin
property_key: owner_corrects_wrong_isin_superseding_transactions
canonical_key: mcp-kibi-model-requirement:ingest.ticker_isin:owner_corrects_wrong_isin_superseding_transactions:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C988B01E8BE311E8
claim_text: The owner must be able to correct a ticker's wrong confirmed ISIN, which supersedes the transactions it booked.
origin:
  kind: agent
  recorded_at: '2026-11-01T12:29:47.316Z'
id: FACT-PROP-1A040AFA587F72EA
type: fact
---
