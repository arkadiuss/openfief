---
title: Removes Instance Keeps Data = true
status: active
text_ref: docs/DESIGN.md#D-58
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-58
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.uninstall_cli
property_key: removes_instance_keeps_data
canonical_key: mcp-kibi-model-requirement:plugins.uninstall_cli:removes_instance_keeps_data:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-013BF51FDA453CC8
claim_text: openfief plugin uninstall must delete the instance's PluginGrant, its Plugin, whose workload the cluster garbage-collects and whose egress allow-list entry the operator drops, and the Secrets no other instance names, and must never touch the data the instance brought into the ledger.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:35.197Z'
id: FACT-PROP-5ED2742CCA4A8C63
type: fact
---
