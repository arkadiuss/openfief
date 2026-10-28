---
title: The operator must deploy a plugin image only when an image verifier confirms its signature, and until cosign verification exists it must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
status: open
text_ref: docs/DESIGN.md#D-17
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-17
  - lane:strict
semantic_text: The operator must deploy a plugin image only when an image verifier confirms its signature, and until cosign verification exists it must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
logic_claims:
  - CLAIM-89F8C266C1C6BC69
semantic_clauses:
  - The operator must deploy a plugin image only when an image verifier confirms its signature, and until cosign verification exists it must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f95bf8deffb0c47f9186bae19744d47a50f8178d0df88e38749bbd8d2845db9c
semantic_inventory:
  - claim_key: CLAIM-89F8C266C1C6BC69
    claim_text: The operator must deploy a plugin image only when an image verifier confirms its signature, and until cosign verification exists it must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented
    role: normative
    status: modeled
    span:
      start: 0
      end: 226
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:23.617Z'
id: REQ-operator-signature-required
type: req
---
The operator must deploy a plugin image only when an image verifier confirms its signature, and until cosign verification exists it must deploy no plugin image and report SignatureVerified as Unknown with reason NotImplemented.

## Context

D-17 says plugin images are verified with cosign from M0. Verification (sigstore trust root, signer identity policy, fetching signatures from the registry) is not built yet, so the operator has a verifier seam whose only implementation verifies nothing and blocks every deployment, rather than one that pretends success. Refines REQ-AUTO-AC7E986366466A6A. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> Plugin images must be pinned by digest, and their signatures are verified with cosign.

docs/DESIGN.md D-17
