---
title: Operator Gates Plugins = true
status: active
text_ref: design-decisions:D-200#11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-11
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: operator_gates_plugins
canonical_key: mcp-kibi-model-requirement:m0.exit:operator_gates_plugins:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5B6D235B2E3277CA
claim_text: A Plugin without a matching PluginGrant must not be deployed, an upgrade requesting a new egress host must stay on the old version until granted, an image referenced by tag or without a valid cosign signature from a trusted signer must be rejected, including when applied with plain kubectl, and the Plugin's status must show verified only for an image the catalog pins.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:32:12.993Z'
id: FACT-PROP-A23DEC2F1092FB71
type: fact
---
