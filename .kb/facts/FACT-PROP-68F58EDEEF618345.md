---
title: Runs Only Under Grant = true
status: active
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: operator.plugin_reconcile
property_key: runs_only_under_grant
canonical_key: mcp-kibi-model-requirement:operator.plugin_reconcile:runs_only_under_grant:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E2F84D82D3E62780
claim_text: The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:05.489Z'
id: FACT-PROP-68F58EDEEF618345
type: fact
---
