---
title: Merge Selection = diff_against_first_parent
status: active
text_ref: docs/dev/sdlc.md#gate-tiers
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-gate-tiers
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.pre_commit_hook
property_key: merge_selection
canonical_key: docs-dev-sdlc-md:tooling.pre_commit_hook:merge_selection:eq:diff_against_first_parent
operator: eq
value_type: string
value_string: diff_against_first_parent
claim_key: CLAIM-D7A50B37DFE77C01
claim_text: The pre-commit hook must check a merge being committed from its diff against its first parent with the same affected-check selection as any other commit, so a merge of a test-only branch runs that branch's checks and not the whole check.
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:41.392Z'
id: FACT-PROP-4CDA2AE51057FB70
type: fact
---
