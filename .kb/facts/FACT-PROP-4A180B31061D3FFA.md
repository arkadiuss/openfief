---
title: Reverse Proxies View Plugin Paths = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-81-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-81#2
fact_kind: property_value
subject_key: core.browser_entry
property_key: reverse_proxies_view_plugin_paths
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:core.browser_entry:reverse_proxies_view_plugin_paths:eq:true
claim_key: CLAIM-791327F46736971D
claim_text: The core must reverse-proxy the paths a browser-facing view plugin serves, such as the built-in UI's, to that plugin, so that only the core ever calls a plugin.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:11:57.596Z'
id: FACT-PROP-4A180B31061D3FFA
type: fact
---
