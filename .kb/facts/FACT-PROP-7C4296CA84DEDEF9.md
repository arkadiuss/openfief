---
title: Carries Core Auth Material = true
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.workload
property_key: carries_core_auth_material
canonical_key: mcp-kibi-model-requirement:plugins.workload:carries_core_auth_material:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A799BE6B51BB4F8F
claim_text: Every plugin's workload must mount the ConfigMap openfief-sa-jwks read-only and set OPENFIEF_INSTANCE to the instance's name, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, so the plugin can verify the core's token without Kubernetes API access.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:45.597Z'
id: FACT-PROP-7C4296CA84DEDEF9
type: fact
---
