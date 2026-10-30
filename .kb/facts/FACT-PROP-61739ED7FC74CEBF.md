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
canonical_key: mcp-kibi-model-requirement:tooling.pre_commit_hook:falls_back_to_full_check:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DD99E75C6415489D
claim_text: The pre-commit hook must run the full check when a change touches an input every check reads, when a merge is being committed, or when the selector cannot place a changed path.
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:44.462Z'
id: FACT-PROP-61739ED7FC74CEBF
type: fact
---
