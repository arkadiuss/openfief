---
title: Books Free Funds Interest As Cash Income = true
status: active
text_ref: docs/DESIGN.md#D-152
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-152
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.xtb_mapper
property_key: books_free_funds_interest_as_cash_income
canonical_key: mcp-kibi-model-requirement:ingest.xtb_mapper:books_free_funds_interest_as_cash_income:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3148369A8B15BB1D
claim_text: The XTB connector must book each free-funds interest row of an xStation export as cash income, with the free-funds interest tax rows of its date as the tax withheld, and report a tax row with no single interest row of its date, or a refund, as an ingest problem.
origin:
  kind: agent
  recorded_at: '2026-11-02T05:51:29.396Z'
id: FACT-PROP-C320A02E9F81DFAB
type: fact
---
