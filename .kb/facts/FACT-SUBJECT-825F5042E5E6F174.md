---
title: Identity Session
status: active
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.session
canonical_key: identity.session
claim_key: CLAIM-680DD6D38E718879
claim_text: After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT.
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:06.137Z'
id: FACT-SUBJECT-825F5042E5E6F174
type: fact
---
