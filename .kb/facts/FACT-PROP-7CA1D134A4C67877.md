---
title: Consumed Fifo = true
status: active
text_ref: docs/DESIGN.md#31-system-of-record-and-views
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-31-system-of-record-and-views
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.lot
property_key: consumed_fifo
canonical_key: mcp-kibi-model-requirement:ledger.lot:consumed_fifo:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E9334EFBE9C77039
claim_text: Disposals must consume an account's lots of an instrument oldest first, by open date and then by the acquiring transaction's key, with an opening lot ordered by its broker-reported open date.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:18.525Z'
id: FACT-PROP-7CA1D134A4C67877
type: fact
---
