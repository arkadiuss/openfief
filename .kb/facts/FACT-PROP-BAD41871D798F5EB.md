---
title: Opening Lots By Open Price = true
status: active
text_ref: docs/DESIGN.md#D-159
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-159
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.xtb_connector
property_key: opening_lots_by_open_price
canonical_key: mcp-kibi-model-requirement:ingest.xtb_connector:opening_lots_by_open_price:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1182E83067890575
claim_text: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position, on the report''s date, with its volume, open date and open price in its trading currency, by ticker until the owner confirms its ISIN.'
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:38.869Z'
id: FACT-PROP-BAD41871D798F5EB
type: fact
---
