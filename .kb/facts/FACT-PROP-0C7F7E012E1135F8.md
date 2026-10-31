---
title: Decision Legs Checked = true
status: active
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.transfer
property_key: decision_legs_checked
canonical_key: mcp-kibi-model-requirement:ledger.transfer:decision_legs_checked:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-048BED1F7B216729
claim_text: A transfer decision must name a stored withdrawal and a stored deposit whose current versions are in different accounts, of exactly the same amount and currency, at most five calendar days apart, and neither opening cash.
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:59.325Z'
id: FACT-PROP-0C7F7E012E1135F8
type: fact
---
