---
title: Plugins Signer Identity
status: active
text_ref: docs/DESIGN.md#D-45
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-45
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.signer_identity
canonical_key: plugins.signer_identity
claim_key: CLAIM-37A68FC72DBF9EAF
claim_text: A signature must count as a trusted signer's only when its certificate has the signer's exact OIDC issuer and either the signer's exact subject or a subject that the signer's pattern matches in full, or when the signer's public key verified it.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:51:12.545Z'
id: FACT-SUBJECT-F68DEFF0C7AFE491
type: fact
---
