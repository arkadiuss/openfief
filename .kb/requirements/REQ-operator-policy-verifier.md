---
title: 'The operator must decide every plugin image with the trust policy and record the outcome in the Plugin''s SignatureVerified condition: True with reason Verified or SignedByTrustedSigner, False with the refusal reason, or Unknown when the registry could not be read, which is retried.'
status: open
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-56
  - lane:strict
semantic_text: 'The operator must decide every plugin image with the trust policy and record the outcome in the Plugin''s SignatureVerified condition: True with reason Verified or SignedByTrustedSigner, False with the refusal reason, or Unknown when the registry could not be read, which is retried.'
logic_claims:
  - CLAIM-80A3EB5FAC5DCC7A
semantic_clauses:
  - 'The operator must decide every plugin image with the trust policy and record the outcome in the Plugin''s SignatureVerified condition: True with reason Verified or SignedByTrustedSigner, False with the refusal reason, or Unknown when the registry could not be read, which is retried.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4effa5526e62e7925715053fb3e9f0acca2ea3c687c89bf49ca56c4ba8f0f09a
semantic_inventory:
  - claim_key: CLAIM-80A3EB5FAC5DCC7A
    claim_text: 'The operator must decide every plugin image with the trust policy and record the outcome in the Plugin''s SignatureVerified condition: True with reason Verified or SignedByTrustedSigner, False with the refusal reason, or Unknown when the registry could not be read, which is retried'
    role: normative
    status: modeled
    span:
      start: 0
      end: 281
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:36.442Z'
id: REQ-operator-policy-verifier
type: req
---
The operator must decide every plugin image with the trust policy and record the outcome in the Plugin's SignatureVerified condition: True with reason Verified or SignedByTrustedSigner, False with the refusal reason, or Unknown when the registry could not be read, which is retried.

## Context

D-56 has the operator check every image itself, including a Plugin applied with kubectl, and record whether it is verified or only signed by a trusted signer, for the UI to show; D-57 never deploys a refused image. A signed image's registry content is cached by digest, so a running plugin is not torn down while its registry is briefly unreachable; refusals are never cached. Refines REQ-plugins-verified-status and REQ-plugins-unsigned-never-deployed. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> It records in the `Plugin`'s status whether the image is verified (D-44) or only signed by a trusted signer, and the UI shows that status.

docs/DESIGN.md D-56
