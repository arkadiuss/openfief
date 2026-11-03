---
title: Job Errors Carry No Values = true
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: valuation.revaluation_jobs
property_key: job_errors_carry_no_values
canonical_key: mcp-kibi-model-requirement:valuation.revaluation_jobs:job_errors_carry_no_values:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-62D71C1661715CD6
claim_text: The revaluation workers must hand River a failed job's error named by the job and by the known errors in its chain, or else its innermost error's type, so that neither River's log nor river_job.errors keeps an amount, a holding, an ISIN or an account number.
origin:
  kind: agent
  recorded_at: '2026-11-03T15:15:54.197Z'
id: FACT-PROP-C526AD95246AAFD6
type: fact
---
