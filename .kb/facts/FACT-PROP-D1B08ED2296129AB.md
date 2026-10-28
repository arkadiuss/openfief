---
title: Fixed Sources No Override = true
status: active
text_ref: docs/DESIGN.md#D-45
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-45
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: operator.trust_config
property_key: fixed_sources_no_override
canonical_key: mcp-kibi-model-requirement:operator.trust_config:fixed_sources_no_override:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9F39D68871C76E2A
claim_text: The operator must build its trust policy from OpenFief's CI identity, the catalog embedded in its binary and an optional owner signer file, must refuse to start when that file is invalid, and must offer no setting that deploys an image the policy refuses.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:42.311Z'
id: FACT-PROP-D1B08ED2296129AB
type: fact
---
