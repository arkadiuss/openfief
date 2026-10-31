---
title: Same Transaction As Change = true
status: active
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: attention.outbox
property_key: same_transaction_as_change
canonical_key: mcp-kibi-model-requirement:attention.outbox:same_transaction_as_change:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E661487AD21DB6E1
claim_text: Each change to an attention item and its outbox event must be written in one database transaction, and the items, their history and the outbox must be append-only.
origin:
  kind: agent
  recorded_at: '2026-10-31T18:31:59.407Z'
id: FACT-PROP-BBFDB0F1DB21C7F4
type: fact
---
