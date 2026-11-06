---
title: Ui Service Worker
status: active
text_ref: docs/DESIGN.md#D-151
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-151
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ui.service_worker
canonical_key: ui.service_worker
claim_key: CLAIM-3C23A7DAE2E8C42A
claim_text: The UI's service worker must cache only static files (styles, scripts, fonts, icons, the manifest and a static offline page) and never a page, fragment or form response; it must empty its cache and delete the kept overview on sign-out, drop its cache on a new version, and show the offline page when the network is gone, with the overview kept for offline reading on the overview's address and nothing of the record elsewhere.
origin:
  kind: agent
  recorded_at: '2026-11-02T10:11:05.110Z'
id: FACT-SUBJECT-AE65AD2A1F920E3F
type: fact
---
