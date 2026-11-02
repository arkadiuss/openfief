---
title: Exact Averages Whole Zloty Tax = true
status: active
text_ref: docs/DESIGN.md#D-144
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-144
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: pl.oki_yearly_tax
property_key: exact_averages_whole_zloty_tax
canonical_key: mcp-kibi-model-requirement:pl.oki_yearly_tax:exact_averages_whole_zloty_tax:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-917B5329BF20EBE4
claim_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, kept exact, exempting savings assets first then investment assets up to the year's amounts, rounding the rest and its tax at the year's rate to whole złoty, under 50 groszy down, and marking a year counted through a day before its last as an estimate, with no rate or tax while the minister hasn't announced the year's rate or exempt amounts.
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:42.068Z'
id: FACT-PROP-3C35D4EE9A982B68
type: fact
---
