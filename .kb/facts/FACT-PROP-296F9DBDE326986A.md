---
title: Has Limit Range = true
status: active
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: deploy.openfief_namespace
property_key: has_limit_range
canonical_key: mcp-kibi-model-requirement:deploy.openfief_namespace:has_limit_range:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-27558F20D5AC5F1A
claim_text: The install's cluster-admin base must give the openfief namespace a LimitRange that sets a default CPU and memory request and a default memory limit for every container that sets none of its own.
origin:
  kind: agent
  recorded_at: '2026-11-02T20:55:48.755Z'
id: FACT-PROP-296F9DBDE326986A
type: fact
---
