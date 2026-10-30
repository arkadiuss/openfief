---
title: Skips E2e And Pg For Docs Only Prs = true
status: active
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.ci_jobs
property_key: skips_e2e_and_pg_for_docs_only_prs
canonical_key: mcp-kibi-model-requirement:tooling.ci_jobs:skips_e2e_and_pg_for_docs_only_prs:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-73EAFF8824A5EF25
claim_text: CI must skip the end-to-end and PostgreSQL jobs for a pull request that changes only documentation or requirements, and must run them on every push to main.
origin:
  kind: agent
  recorded_at: '2026-10-30T03:34:18.369Z'
id: FACT-PROP-4C68A3E8854EFE68
type: fact
---
