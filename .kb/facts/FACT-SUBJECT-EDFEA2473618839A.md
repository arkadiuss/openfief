---
title: Demo Deploy
status: active
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: demo.deploy
canonical_key: demo.deploy
claim_key: CLAIM-E97064649F043A4A
claim_text: Demo deploys must pin each image in deploy/demo to the digest whose keyless cosign signature from this repository's images workflow on main verifies, reach the cluster only through a pull request that Flux pulls once merged, and keep the OpenFief layer suspended while any digest is a placeholder.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:15.444Z'
id: FACT-SUBJECT-EDFEA2473618839A
type: fact
---
