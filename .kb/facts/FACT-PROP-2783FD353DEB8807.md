---
title: Caches Static Files And Shows Kept Overview = true
status: active
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ui.service_worker
property_key: caches_static_files_and_shows_kept_overview
canonical_key: mcp-kibi-model-requirement:ui.service_worker:caches_static_files_and_shows_kept_overview:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3C23A7DAE2E8C42A
claim_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere.
origin:
  kind: agent
  recorded_at: '2026-11-02T10:11:07.576Z'
id: FACT-PROP-2783FD353DEB8807
type: fact
---
