---
title: Decides Verified Unverified Or Refused = true
status: active
text_ref: docs/DESIGN.md#D-44
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-44
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.trust_policy
property_key: decides_verified_unverified_or_refused
canonical_key: mcp-kibi-model-requirement:plugins.trust_policy:decides_verified_unverified_or_refused:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8FB52869CF7C6420
claim_text: The trust policy must call an image verified only when a catalog entry pins its digest and that entry's publisher signed it, unverified when any other trusted signer signed it, and must refuse it otherwise.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:51:41.105Z'
id: FACT-PROP-4B58CF4AADD27E53
type: fact
---
