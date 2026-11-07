---
title: Cash Within Cent Holding Exact = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.balance_check
property_key: cash_within_cent_holding_exact
canonical_key: mcp-kibi-model-requirement:views.balance_check:cash_within_cent_holding_exact:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-206B121C6E083E44
claim_text: 'The owner''s balance check must show, for each balance assertion, the reported balance, the balance derived from the transaction log, reported minus derived, and whether it passes: a cash assertion within 0.01 of its currency either way, a holding assertion only with equal units.'
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:19.496Z'
id: FACT-PROP-D6363303FCD8755B
type: fact
---
