---
title: Books Under Provisional Isin = true
status: active
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.treasury_bonds
property_key: books_under_provisional_isin
canonical_key: mcp-kibi-model-requirement:plugins.treasury_bonds:books_under_provisional_isin:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-15B1643C9851CF3C
claim_text: Until the ministry's data lists a series' ISIN, or while it can't be read and no lot needs its rates, the treasury bond connector must book the series' lots under its provisional ISIN, PL, the series code, 00 and a check digit, keyed by series and purchase date as under the registry's ISIN.
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:06.953Z'
id: FACT-PROP-1FE1931BC1681E2F
type: fact
---
