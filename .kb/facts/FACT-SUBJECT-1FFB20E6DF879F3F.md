---
title: Identity Script Token
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.script_token
canonical_key: identity.script_token
claim_key: CLAIM-14EF7DB3875CDB23
claim_text: Issuing a script token must create a read-scoped token with a valid name and a lifetime above zero and at most one year, from random bytes of crypto/rand, return its secret only once, and store only the SHA-256 hash of the secret's key.
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:18.183Z'
id: FACT-SUBJECT-1FFB20E6DF879F3F
type: fact
---
