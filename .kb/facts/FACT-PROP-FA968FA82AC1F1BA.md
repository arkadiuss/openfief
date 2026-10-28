---
title: Atomic With Hooks = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.import
property_key: atomic_with_hooks
canonical_key: mcp-kibi-model-requirement:ledger.import:atomic_with_hooks:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-ACF447E69493DF9D
claim_text: An import must append its transactions, store its assertion checks and run its import hooks in one database transaction that commits whole or not at all, and must report per account the earliest date whose current transactions it changed.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:18.029Z'
id: FACT-PROP-FA968FA82AC1F1BA
type: fact
---
