---
title: Optional Secret Mounted Optional = true
status: active
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.manifest_secret
property_key: optional_secret_mounted_optional
canonical_key: mcp-kibi-model-requirement:plugins.manifest_secret:optional_secret_mounted_optional:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E2AAB6F33C41BABF
claim_text: A plugin manifest may mark a Secret role optional; the operator must mount that Secret as an optional volume so the plugin pod starts while the owner has not created it, every other Secret must stay required, the flag must survive the Plugin resource the CLI writes and reads, and the install summary must say the plugin runs without it.
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:04.181Z'
id: FACT-PROP-A34D8A34CDA4656C
type: fact
---
