---
title: Refuses Non Owner And Malformed = true
status: active
text_ref: docs/DESIGN.md#D-158
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-158
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.cash_checkpoint_service
property_key: refuses_non_owner_and_malformed
canonical_key: mcp-kibi-model-requirement:ledger.cash_checkpoint_service:refuses_non_owner_and_malformed:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-72EE8F3A3344A2BE
claim_text: The CashCheckpointService must admit only the owner's view, refuse an account the owner may not change, and refuse a malformed date, currency or amount, an unknown account or checkpoint, and a checkpoint with nothing to book, storing nothing then.
origin:
  kind: agent
  recorded_at: '2026-11-02T03:50:15.070Z'
id: FACT-PROP-E32BF6890805D98E
type: fact
---
