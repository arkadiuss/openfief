---
title: Sets Only Core Forwarding Headers = true
status: active
text_ref: docs/DESIGN.md#D-81
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-81
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.browser_proxy
property_key: sets_only_core_forwarding_headers
canonical_key: mcp-kibi-model-requirement:views.browser_proxy:sets_only_core_forwarding_headers:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-14EE12207C3EB2E9
claim_text: When the core proxies a browser request to a view plugin, it must present its own token for that instance and set the Host, X-Forwarded-Proto and X-Forwarded-Host itself, dropping every forwarding header the client sent.
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:09.312Z'
id: FACT-PROP-9540BE3755B3AFC2
type: fact
---
