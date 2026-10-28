---
title: Ensures Defaults Idempotently = true
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
property_key: ensures_defaults_idempotently
canonical_key: mcp-kibi-model-requirement:ledger.sleeve_registry:ensures_defaults_idempotently:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F5D735EC1994D2DF
claim_text: Ensuring the defaults must store every missing built-in sleeve kind and default sleeve in one database transaction, never change a stored kind or sleeve, and store nothing more when run again or by several processes at once.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:53:22.592Z'
id: FACT-PROP-EC400A0B2F140870
type: fact
---
