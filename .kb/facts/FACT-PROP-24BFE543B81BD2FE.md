---
title: Withdraws Paid Out Payouts = true
status: active
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: treasurybonds.payouts
property_key: withdraws_paid_out_payouts
canonical_key: mcp-kibi-model-requirement:treasurybonds.payouts:withdraws_paid_out_payouts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E830DF62E9C06092
claim_text: When an account's payout setting pays out what fell due on a coupon's or redemption's day, the treasury-bonds connector must follow it with a withdrawal of the cash it brought on the same day, keyed after it.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:30.338Z'
id: FACT-PROP-24BFE543B81BD2FE
type: fact
---
