---
title: Postings Sum To Zero Per Currency = true
status: active
text_ref: docs/DESIGN.md#D-19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-19
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.transaction
property_key: postings_sum_to_zero_per_currency
canonical_key: mcp-kibi-model-requirement:ledger.transaction:postings_sum_to_zero_per_currency:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-91ED7AB82979DBB9
claim_text: A transaction must have at least two postings that sum to zero in every currency, and an unbalanced transaction must be rejected with the currency and amount it is off by.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:43.318Z'
id: FACT-PROP-5504DCBAC9B29DC8
type: fact
---
