---
title: Rebooks Payouts After Change Date = true
status: active
text_ref: design-decisions:D-162#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-162-3
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.payout_sync
property_key: rebooks_payouts_after_change_date
canonical_key: mcp-kibi-model-requirement:ingest.payout_sync:rebooks_payouts_after_change_date:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B2544C5D2AF4F210
claim_text: After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones.
origin:
  kind: agent
  recorded_at: '2026-11-02T00:55:26.452Z'
id: FACT-PROP-AC98FE1AAFE0A849
type: fact
---
