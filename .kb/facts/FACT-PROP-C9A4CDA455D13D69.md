---
title: Identifies Plugin By Token = true
status: active
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: identifies_plugin_by_token
canonical_key: mcp-kibi-model-requirement:plugins.egress_proxy:identifies_plugin_by_token:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8A8941C5C74DCF70
claim_text: The egress proxy must identify a plugin only from a projected token sent as the password of a Basic Proxy-Authorization header, which the TokenReview API authenticates for audience openfief-egress as a plugin ServiceAccount, refuse every other request, and never log the token.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:10:12.232Z'
id: FACT-PROP-C9A4CDA455D13D69
type: fact
---
