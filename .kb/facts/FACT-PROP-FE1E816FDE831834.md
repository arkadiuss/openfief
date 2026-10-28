---
title: One Job Per Account = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.river_queue
property_key: one_job_per_account
canonical_key: mcp-kibi-model-requirement:valuation.river_queue:one_job_per_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8CBD69ADFDF487D2
claim_text: Revaluations must run as River jobs, at most one per account waiting or running, with the account's earliest pending date kept beside them and lowered by every enqueue, and a job must complete only in the transaction that clears that date, so that no change is left without a job.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:57:37.825Z'
id: FACT-PROP-FE1E816FDE831834
type: fact
---
