---
title: Mounted Only When Expected = true
status: active
text_ref: design-session:connector-endpoint-ca-bundle-cli
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca-bundle-cli
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.extra_ca
property_key: mounted_only_when_expected
canonical_key: mcp-kibi-model-requirement:plugins.extra_ca:mounted_only_when_expected:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0472AE5A88452CC9
claim_text: A plugin instance must mount its optional ca-bundle Secret only when its explicit extraCa instance setting says an extra CA bundle is expected, which the CLI sets with install --extra-ca and toggles with upgrade --extra-ca or --no-extra-ca, which needs a manifest that asks for the ca-bundle role, and which changes nothing the grant approves.
origin:
  kind: agent
  recorded_at: '2026-11-07T10:23:04.920Z'
id: FACT-PROP-AE2AB0F5B15D2794
type: fact
---
