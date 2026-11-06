---
title: Books Closed Cfd Result As Cash Income = true
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
property_key: books_closed_cfd_result_as_cash_income
canonical_key: mcp-kibi-model-requirement:ingest.xtb_mapper:books_closed_cfd_result_as_cash_income:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3B0BD8DCE3FEC389
claim_text: The XTB connector must book each closed CFD position's Profit/Loss as cash income on the UTC date it closed, keyed by its position ID, and report every open CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
origin:
  kind: agent
  recorded_at: '2026-11-02T05:51:36.149Z'
id: FACT-PROP-DD6A74FBC7CCA26F
type: fact
---
