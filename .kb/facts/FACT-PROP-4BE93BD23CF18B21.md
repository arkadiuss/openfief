---
title: Falls Back To Full Check = true
status: active
text_ref: docs/dev/sdlc.md#tooling
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-tooling
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.pre_commit_hook
property_key: falls_back_to_full_check
canonical_key: docs-dev-sdlc-md:tooling.pre_commit_hook:falls_back_to_full_check:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5617E5D04B6C944C
claim_text: The pre-commit hook must run the full check when a change touches an input every check reads or when the selector cannot place a changed path, and must plan a merge being committed from its diff against its first parent like any other commit.
origin:
  kind: agent
  recorded_at: '2026-11-12T15:12:07.711Z'
id: FACT-PROP-4BE93BD23CF18B21
type: fact
---
