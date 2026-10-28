---
title: Manifests Validated In Check = true
status: active
text_ref: docs/dev/sdlc.md#flow
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-flow
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.deploy
property_key: manifests_validated_in_check
canonical_key: mcp-kibi-model-requirement:demo.deploy:manifests_validated_in_check:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1B72CE74D4BB0BE3
claim_text: task check must build deploy/demo, every git path its Flux Kustomizations apply and the deploy/demo/openfief template with the pinned kustomize, offline, and must fail when an image is not pinned by a sha256 digest or an OCI source does not require a cosign signature.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:01:30.209Z'
id: FACT-PROP-D7FAECC6B61FCBE6
type: fact
---
