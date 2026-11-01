---
title: Release Signatures
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: release.signatures
canonical_key: release.signatures
claim_key: CLAIM-115A85ED364C01FA
claim_text: Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify.
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:43.339Z'
id: FACT-SUBJECT-6A67E472A68A30B4
type: fact
---
