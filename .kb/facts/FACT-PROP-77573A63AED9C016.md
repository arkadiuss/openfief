---
title: Service Worker Caches Static Only = true
status: active
text_ref: docs/DESIGN.md#D-69
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-69
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ui
property_key: service_worker_caches_static_only
canonical_key: mcp-kibi-model-requirement:views.ui:service_worker_caches_static_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8CA1CB1405056DD7
claim_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache on sign-out and drop it on a new version, and show the offline page, with no data, when the network is gone.
origin:
  kind: agent
  recorded_at: '2026-10-30T08:08:20.253Z'
id: FACT-PROP-77573A63AED9C016
type: fact
---
