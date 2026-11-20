---
title: Plugins Signature Trust
status: active
text_ref: design-decisions:D-169#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-169-1
  - lane:strict
  - fact:subject
fact_kind: subject
subject_key: plugins.signature_trust
canonical_key: plugins.signature_trust
claim_key: CLAIM-D1B8E9C11D1FEDD6
claim_text: When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature.
origin:
  kind: agent
  recorded_at: '2026-11-01T21:51:18.568Z'
id: FACT-SUBJECT-9AFA7913446842D1
type: fact
---

Definition not stated: the subject plugins.signature_trust is named by requirement REQ-plugins-untrusted-signer-named, which links to the subject fact FACT-SUBJECT-9AFA7913446842D1 through constrains. The people who own that source can say what it is in the product and what it is not.
