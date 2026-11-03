---
title: Security Headers Set = true
status: active
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.front_door
property_key: security_headers_set
canonical_key: mcp-kibi-model-requirement:views.front_door:security_headers_set:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CFCA1C09E8202DF1
claim_text: Every answer the core proxies from a view plugin, and the front door's own refusals, must carry the core's content security policy next to any the view sends, X-Frame-Options DENY, X-Content-Type-Options nosniff and Referrer-Policy no-referrer.
origin:
  kind: agent
  recorded_at: '2026-11-03T00:11:01.662Z'
id: FACT-PROP-6EEE827D19578C7B
type: fact
---
