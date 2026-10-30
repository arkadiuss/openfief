---
title: Decision Versioned = true
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
property_key: decision_versioned
canonical_key: mcp-kibi-model-requirement:ledger.transfer:decision_versioned:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9C5D6D1E56CCAE07
claim_text: A decision on a transfer pair that differs from its current one must be appended as the pair's next version, superseding the earlier ones, and repeating the current decision must append nothing.
origin:
  kind: agent
  recorded_at: '2026-10-30T06:06:15.296Z'
id: FACT-PROP-2D8AED09C8644D9D
type: fact
---
