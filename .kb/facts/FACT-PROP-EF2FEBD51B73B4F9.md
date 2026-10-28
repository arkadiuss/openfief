---
title: Has Minimal Identity And Granted Secrets = true
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.workload
property_key: has_minimal_identity_and_granted_secrets
canonical_key: mcp-kibi-model-requirement:plugins.workload:has_minimal_identity_and_granted_secrets:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0997891548AF7E46
claim_text: A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:31.973Z'
id: FACT-PROP-EF2FEBD51B73B4F9
type: fact
---
