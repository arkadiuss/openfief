---
title: Checks Sleeves Against Stored Kinds = true
status: active
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.sleeve_registry
property_key: checks_sleeves_against_stored_kinds
canonical_key: mcp-kibi-model-requirement:ledger.sleeve_registry:checks_sleeves_against_stored_kinds:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-600720DA21BCDE78
claim_text: The ledger must check every sleeve kind and sleeve it saves against the stored kinds, store a sleeve with no role with its kind's default role, and refuse a sleeve with an unknown kind, a role its kind does not allow, or a change of a stored kind's default role.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:53:13.713Z'
id: FACT-PROP-A1717433E7CD0FB7
type: fact
---
