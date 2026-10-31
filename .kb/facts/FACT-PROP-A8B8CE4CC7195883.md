---
title: Broken Confirmed Transfer Raises Attention Item = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-131-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-131#2
fact_kind: property_value
subject_key: ledger.transfer
property_key: broken_confirmed_transfer_raises_attention_item
operator: eq
value_type: bool
value_bool: true
canonical_key: kb-requirements-req-ledger-transfer-broken-attention-item-md:ledger.transfer:broken_confirmed_transfer_raises_attention_item:eq:true
claim_key: CLAIM-0C626D39ED5F47D1
claim_text: A confirmed transfer whose legs no longer match after a correction must raise an attention item, so the owner confirms the corrected pair or rejects it.
origin:
  kind: agent
  recorded_at: '2026-10-31T14:46:15.193Z'
id: FACT-PROP-A8B8CE4CC7195883
type: fact
---
