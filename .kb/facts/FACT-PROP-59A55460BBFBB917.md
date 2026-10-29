---
title: Keeps Instance Grants New Request = true
status: active
text_ref: docs/DESIGN.md#D-90
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-90
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.cli_upgrade
property_key: keeps_instance_grants_new_request
canonical_key: mcp-kibi-model-requirement:plugins.cli_upgrade:keeps_instance_grants_new_request:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-BF35988B00FCFA12
claim_text: An upgrade must keep the instance's name, its config unless the owner gives a new one, and its in-cluster service while the new manifest asks for one, grant exactly what the new version requests, and report what that adds and removes.
origin:
  kind: agent
  recorded_at: '2026-10-29T17:04:13.220Z'
id: FACT-PROP-59A55460BBFBB917
type: fact
---
