---
title: Routes To The One Browser View = true
status: active
text_ref: docs/DESIGN.md#D-114
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-114
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.browser_routing
property_key: routes_to_the_one_browser_view
canonical_key: mcp-kibi-model-requirement:views.browser_routing:routes_to_the_one_browser_view:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3BEFBA00B96C998F
claim_text: The core must send a browser path it does not serve to the one view plugin instance holding a granted browser path when that path covers it by whole segments, answer 404 when it does not or no instance holds one, and serve none while more than one instance holds a browser path.
origin:
  kind: agent
  recorded_at: '2026-10-29T20:58:35.831Z'
id: FACT-PROP-AFC61D0296F4ECBD
type: fact
---
