---
title: Verified As Install Guide = true
status: active
text_ref: docs/dev/sdlc.md#releases
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-releases
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: release.signatures
property_key: verified_as_install_guide
canonical_key: mcp-kibi-model-requirement:release.signatures:verified_as_install_guide:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-115A85ED364C01FA
claim_text: Every release run, a dry run included, must verify with cosign the CLI files it signed as release.yml at the run's ref and each image its install runs as images.yml on main, issued by GitHub Actions, and fail on any file or image that does not verify.
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:45.932Z'
id: FACT-PROP-C9E21D0EF63BD383
type: fact
---
