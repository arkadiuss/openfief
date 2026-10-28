---
title: Checked Before Merge = true
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.proto_contracts
property_key: checked_before_merge
canonical_key: mcp-kibi-model-requirement:tooling.proto_contracts:checked_before_merge:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-314DA7F5493D86DD
claim_text: Changes to the protobuf contracts must pass buf lint with the STANDARD rules and buf breaking against the base revision, and the committed generated Go code must equal what the pinned generators produce from them.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:37:27.954Z'
id: FACT-PROP-0280E88A61B5AC54
type: fact
---
