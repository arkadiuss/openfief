---
title: Identity Passkey Enrollment
status: active
text_ref: design-decisions:D-60#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-60-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: identity.passkey_enrollment
canonical_key: identity.passkey_enrollment
claim_key: CLAIM-7DD67F5DF8CCE695
claim_text: OpenFief must install with no hostname, and the owner's first passkey must be created on localhost through a one-time link the CLI prints.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:42.198Z'
id: FACT-SUBJECT-9565864B05E4ACEA
type: fact
---

Definition not stated: the subject identity.passkey_enrollment is named by requirements REQ-identity-core-refuses-enroll-origin, REQ-identity-enroll-localhost-forward, REQ-identity-enroll-origin, REQ-identity-first-passkey-command, REQ-identity-first-passkey-link, REQ-identity-first-passkey-own-host-or-localhost, REQ-identity-first-run-wizard-follows, REQ-identity-install-without-hostname, REQ-identity-passkey-new-hostname, which link to the subject fact FACT-SUBJECT-9565864B05E4ACEA through constrains. The people who own that source can say what it is in the product and what it is not.
