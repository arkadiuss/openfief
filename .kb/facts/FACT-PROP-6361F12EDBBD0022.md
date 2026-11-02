---
title: Earlier Payments At Go Live = true
status: active
text_ref: docs/DESIGN.md#D-149
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-149
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ike_paid_in
property_key: earlier_payments_at_go_live
canonical_key: mcp-kibi-model-requirement:views.ike_paid_in:earlier_payments_at_go_live:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-354B0EF0531AE968
claim_text: What was paid into an IKE seeded from a starting balance must be the earlier payments the owner entered, dated on its go-live, plus its later deposits less its withdrawals, its seeding being no payment, and its liquidation total must be missing until the earlier payments are entered.
origin:
  kind: agent
  recorded_at: '2026-11-01T20:13:11.472Z'
id: FACT-PROP-6361F12EDBBD0022
type: fact
---
