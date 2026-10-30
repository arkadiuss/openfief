---
title: Carries Signed Digest Forward = true
status: active
text_ref: docs/dev/sdlc.md#ci-and-environments
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-sdlc-md-ci-and-environments
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: tooling.ci_images
property_key: carries_signed_digest_forward
canonical_key: mcp-kibi-model-requirement:tooling.ci_images:carries_signed_digest_forward:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EA5E2E006EC04D62
claim_text: An image CI does not rebuild on main must keep the digest the images workflow on main signed for an earlier commit, tagged for the new commit only after that signature verifies, so the demo deploy finds a signed digest for every image.
origin:
  kind: agent
  recorded_at: '2026-10-30T03:34:07.092Z'
id: FACT-PROP-DECAF2E34B829511
type: fact
---
