---
title: Identity Core Token
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
subject_key: identity.core_token
canonical_key: identity.core_token
claim_key: CLAIM-5C8E0FE7BAC6701B
claim_text: The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:43.363Z'
id: FACT-SUBJECT-D2FDAD9362808F55
type: fact
---
