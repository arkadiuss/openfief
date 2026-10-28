---
title: Pins Verified Versions = true
status: active
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.cloud_init
property_key: pins_verified_versions
canonical_key: mcp-kibi-model-requirement:demo.cloud_init:pins_verified_versions:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-DB9445E3CB5F690B
claim_text: A demo server's cloud-init must install k3s and Flux at the versions pinned in the Taskfile, each download checked against a pinned sha256, with Traefik and servicelb disabled and no secrets, and rendering must fail when a pinned value is missing or malformed or a placeholder is left.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:06.037Z'
id: FACT-PROP-52D59A932D10D5AE
type: fact
---
