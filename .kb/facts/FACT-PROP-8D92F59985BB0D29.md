---
title: Deterministic For Same Authored Kb = true
status: active
text_ref: docs/dev/pipeline.md#the-kibi-gate
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-pipeline-md-the-kibi-gate
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.kibi_gate
property_key: deterministic_for_same_authored_kb
canonical_key: mcp-kibi-model-requirement:tooling.kibi_gate:deterministic_for_same_authored_kb:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-708AC38A1DCF5BF9
claim_text: 'The full kibi gate must give the same answer for the same authored knowledge base: a failure that names a kibi compile defect (an authored relationship missing from the compiled KB, or a duplicate entity row in the paginated projection) must recompile the store once from the same files and let the second check decide, and any other failure must fail at once.'
origin:
  kind: agent
  recorded_at: '2026-10-30T04:09:52.094Z'
id: FACT-PROP-8D92F59985BB0D29
type: fact
---
