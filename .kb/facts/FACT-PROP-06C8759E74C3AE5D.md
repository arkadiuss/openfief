---
title: Routes By Longest Granted Prefix = true
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
subject_key: views.browser_routing
property_key: routes_by_longest_granted_prefix
canonical_key: mcp-kibi-model-requirement:views.browser_routing:routes_by_longest_granted_prefix:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E7C8121653F304DA
claim_text: The core must send a browser path it does not serve to the running view plugin whose granted browser path is the longest prefix covering it by whole segments, answer 404 when none covers it, and serve neither when two plugins claim that prefix.
origin:
  kind: agent
  recorded_at: '2026-10-29T12:02:19.400Z'
id: FACT-PROP-06C8759E74C3AE5D
type: fact
---
