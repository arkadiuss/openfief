---
title: Added One At A Time On Go Live = true
status: active
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.opening_lots
property_key: added_one_at_a_time_on_go_live
canonical_key: mcp-kibi-model-requirement:ledger.opening_lots:added_one_at_a_time_on_go_live:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-18B50858DCCF5E92
claim_text: Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way.
origin:
  kind: agent
  recorded_at: '2026-10-30T05:02:01.462Z'
id: FACT-PROP-0F0742369E6DBB14
type: fact
---
