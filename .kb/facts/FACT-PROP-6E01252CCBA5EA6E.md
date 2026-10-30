---
title: Only Confirmed Transfers Skip Cash Flows = true
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
subject_key: views.performance
property_key: only_confirmed_transfers_skip_cash_flows
canonical_key: mcp-kibi-model-requirement:views.performance:only_confirmed_transfers_skip_cash_flows:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E94EA8621E2FA705
claim_text: The performance series must leave out of its cash flows only the withdrawal and the deposit of each transfer the owner confirmed, counting the money in transit between them once, while a suggested or rejected pair stays two cash flows.
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:17.671Z'
id: FACT-PROP-6E01252CCBA5EA6E
type: fact
---
