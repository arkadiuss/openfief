---
title: Identity Passkey Check
status: active
text_ref: docs/DESIGN.md#D-183
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-183
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.passkey_check
canonical_key: identity.passkey_check
claim_key: CLAIM-C484F2B18928522F
claim_text: An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again.
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:39.588Z'
id: FACT-SUBJECT-F7C6F52CD03FE0B7
type: fact
---
