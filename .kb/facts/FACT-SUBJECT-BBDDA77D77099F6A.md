---
title: Demo Cloud Init
status: active
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: demo.cloud_init
canonical_key: demo.cloud_init
claim_key: CLAIM-DB9445E3CB5F690B
claim_text: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:01.373Z'
id: FACT-SUBJECT-BBDDA77D77099F6A
type: fact
---
