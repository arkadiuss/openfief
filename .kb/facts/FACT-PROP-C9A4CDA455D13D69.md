---
title: Identifies Plugin By Token = true
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
subject_key: plugins.egress_proxy
property_key: identifies_plugin_by_token
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:identifies_plugin_by_token:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C749997F9CF3C2F0
claim_text: The egress proxy must identify a plugin only from a Proxy-Authorization Bearer projected token that the TokenReview API authenticates for audience openfief-core as a plugin ServiceAccount, refuse every other request, and never log the token.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:12.232Z'
id: FACT-PROP-C9A4CDA455D13D69
type: fact
---
