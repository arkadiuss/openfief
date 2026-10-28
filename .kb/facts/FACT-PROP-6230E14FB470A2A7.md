---
title: Grants Exactly The Signed Request = true
status: active
text_ref: docs/DESIGN.md#D-46
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-46
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.install_cli
property_key: grants_exactly_the_signed_request
canonical_key: mcp-kibi-model-requirement:plugins.install_cli:grants_exactly_the_signed_request:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C3FDC202440DDEE4
claim_text: An install must apply a Plugin that requests exactly the image's signed manifest label with the owner's config, and a PluginGrant of the same name that approves exactly those requested permissions, and must refuse an instance name the Plugin CRD would refuse.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:24:07.458Z'
id: FACT-PROP-6230E14FB470A2A7
type: fact
---
