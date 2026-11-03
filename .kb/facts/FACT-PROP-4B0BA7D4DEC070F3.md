---
title: Sets Own Requests And Memory Limit = true
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
subject_key: deploy.database_cluster
property_key: sets_own_requests_and_memory_limit
canonical_key: mcp-kibi-model-requirement:deploy.database_cluster:sets_own_requests_and_memory_limit:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DA80FFD468ADBB16
claim_text: 'The database Cluster the release bundle ships must set its own resources: 100m CPU and 256Mi of memory requested and a 1Gi memory limit, which the owner sizes.'
origin:
  kind: agent
  recorded_at: '2026-11-03T02:14:45.305Z'
id: FACT-PROP-4B0BA7D4DEC070F3
type: fact
---
