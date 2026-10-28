---
title: The operator must deploy a plugin image only when its image verifier confirms the signature, and an operator built without a verifier must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
status: open
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
semantic_text: The operator must deploy a plugin image only when its image verifier confirms the signature, and an operator built without a verifier must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
logic_claims:
  - CLAIM-5BC1096F3F857964
semantic_clauses:
  - The operator must deploy a plugin image only when its image verifier confirms the signature, and an operator built without a verifier must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5ffc9e8bef5daba311d9bcc8a20337f7c6681ddaf0f415e6cdfcae6cbdf92a6d
semantic_inventory:
  - claim_key: CLAIM-5BC1096F3F857964
    claim_text: The operator must deploy a plugin image only when its image verifier confirms the signature, and an operator built without a verifier must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented
    role: normative
    status: modeled
    span:
      start: 0
      end: 228
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:23.617Z'
id: REQ-operator-signature-required
type: req
---
The operator must deploy a plugin image only when its image verifier confirms the signature, and an operator built without a verifier must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.

## Context

D-17 says plugin images are verified with cosign from M0. The operator keeps a verifier seam: PolicyVerifier is the real implementation that cmd/operator wires in, and UnimplementedVerifier, which verifies nothing and blocks every deployment, remains for tests and as the safe default of the seam rather than one that pretends success. Refines REQ-AUTO-AC7E986366466A6A. Modeled by Claude Code for the M0 signature-verification task; the author has not reviewed it yet.

## Source

> Plugin images must be pinned by digest, and their signatures are verified with cosign.

docs/DESIGN.md D-17
