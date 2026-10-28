---
title: Deploys Verified Digests Automatically = true
status: active
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.deploy
property_key: deploys_verified_digests_automatically
canonical_key: mcp-kibi-model-requirement:demo.deploy:deploys_verified_digests_automatically:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2BED11A9401784F5
claim_text: Every push to main and every beta must deploy to the demo without a manual step, by publishing deploy/demo/openfief pinned to the image digests whose keyless cosign signatures from this repository's images workflow on main verify, as an OCI artifact signed by the deploy workflow, which Flux applies only after verifying that signature.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:05:32.068Z'
id: FACT-PROP-875BB366B951B8AF
type: fact
---
