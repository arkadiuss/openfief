---
title: Skips Go For Docs Only Commits = true
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
property_key: skips_go_for_docs_only_commits
canonical_key: mcp-kibi-model-requirement:tooling.pre_commit_hook:skips_go_for_docs_only_commits:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-505A586F876E963C
claim_text: A commit that changes only documentation or requirements must run no Go step in the pre-commit hook, and must still build the docs, run the docs staleness gate and run kibi's staged check.
origin:
  kind: agent
  recorded_at: '2026-10-30T03:33:33.414Z'
id: FACT-PROP-D194E06B8BF739F0
type: fact
---
