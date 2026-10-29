---
title: Grants Operator Service Read = true
status: active
text_ref: docs/DESIGN.md#D-74
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-74
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.cli_install
property_key: grants_operator_service_read
canonical_key: mcp-kibi-model-requirement:plugins.cli_install:grants_operator_service_read:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3AA9565159341E1B
claim_text: An install with an in-cluster service must check that Service with the owner's kubeconfig and create in its namespace a Role and RoleBinding that let only the operator's ServiceAccount get only that Service, and uninstall must delete them.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:11:18.538Z'
id: FACT-PROP-8AF96E3A39C3D061
type: fact
---
