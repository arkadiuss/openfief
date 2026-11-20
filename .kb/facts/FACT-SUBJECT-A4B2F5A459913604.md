---
title: Identity Auth Service
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
subject_key: identity.auth_service
canonical_key: identity.auth_service
claim_key: CLAIM-DCF87CA2A55E5311
claim_text: The core must offer view plugins an authentication service, openfief.identity.v1.AuthService over Connect RPC, separate from the read-only public API, for passkey registration and login, demo read-only sessions, sessions with a read or owner scope, script tokens and enrollment links.
origin:
  kind: agent
  recorded_at: '2026-10-28T18:34:57.716Z'
id: FACT-SUBJECT-A4B2F5A459913604
type: fact
---

Definition not stated: the subject identity.auth_service is named by requirements REQ-identity-auth-callers, REQ-identity-auth-contract, REQ-identity-auth-not-public, which link to the subject fact FACT-SUBJECT-A4B2F5A459913604 through constrains. The people who own that source can say what it is in the product and what it is not.
