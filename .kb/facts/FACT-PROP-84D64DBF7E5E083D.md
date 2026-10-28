---
title: Publishes Running Plugins Hosts = true
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: operator.allow_list
property_key: publishes_running_plugins_hosts
canonical_key: mcp-kibi-model-requirement:operator.allow_list:publishes_running_plugins_hosts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FC6A5236BABD43AE
claim_text: The operator must publish each running plugin's effective egress hosts under the plugin's key in the openfief-egress-allowlist ConfigMap and remove that key when the plugin stops running or is deleted.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:28.890Z'
id: FACT-PROP-84D64DBF7E5E083D
type: fact
---
