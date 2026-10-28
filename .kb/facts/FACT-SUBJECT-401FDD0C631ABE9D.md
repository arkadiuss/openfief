---
title: Sdk Core Verifier
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.core_verifier
canonical_key: sdk.core_verifier
claim_key: CLAIM-D1AEC26052D663F0
claim_text: The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:32.032Z'
id: FACT-SUBJECT-401FDD0C631ABE9D
type: fact
---
