---
title: Isolate Each Plugin = true
status: active
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: operator.plugin_objects
property_key: isolate_each_plugin
canonical_key: mcp-kibi-model-requirement:operator.plugin_objects:isolate_each_plugin:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-58700B0AF172714D
claim_text: The operator must run each plugin as a ServiceAccount without an API token, a Deployment that mounts only its projected token and its granted Secrets read-only, a ClusterIP Service and a NetworkPolicy that denies everything except the compiled rules to the core's and egress proxy's pods.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:31.639Z'
id: FACT-PROP-DB795534EE33A6D2
type: fact
---
