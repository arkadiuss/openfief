---
title: Requires Breaking Commit Mark = true
status: active
text_ref: docs/dev/pipeline.md#breaking-contract-changes
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-breaking-contract-changes
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.breaking_change_exceptions
property_key: requires_breaking_commit_mark
canonical_key: mcp-kibi-model-requirement:tooling.breaking_change_exceptions:requires_breaking_commit_mark:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0B50F30D5BCAF230
claim_text: A commit that adds a breaking-change exception must be marked breaking, with `!` or a BREAKING CHANGE footer, or the commit-msg hook and proto:breaking refuse it.
origin:
  kind: agent
  recorded_at: '2026-11-01T19:21:40.695Z'
id: FACT-PROP-B844DBAF5AA11D54
type: fact
---
