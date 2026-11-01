---
title: Names Release Signing Identities = true
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
subject_key: docs.install_guide
property_key: names_release_signing_identities
canonical_key: mcp-kibi-model-requirement:docs.install_guide:names_release_signing_identities:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A8D4DCE70CDC9D79
claim_text: The install guide must tell the owner to verify the CLI with cosign verify-blob against release.yml on main and the images with cosign verify against images.yml on main, both issued by GitHub Actions, and must name no other identity.
origin:
  kind: agent
  recorded_at: '2026-11-01T14:58:57.993Z'
id: FACT-PROP-760E72B7C6408A4C
type: fact
---
