---
title: When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature.
status: open
text_ref: design-decisions:D-169#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-169-1
  - lane:strict
semantic_text: When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature.
logic_claims:
  - CLAIM-D1B8E9C11D1FEDD6
semantic_clauses:
  - When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1f2226556aeced21bdbc711c1c7d68da81f99ebec5eb7cc6984161f0bc28a306
semantic_inventory:
  - claim_key: CLAIM-D1B8E9C11D1FEDD6
    claim_text: When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature
    role: condition
    status: modeled
    span:
      start: 0
      end: 176
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-169#1
  recorded_at: '2026-11-01T21:51:22.953Z'
id: REQ-plugins-untrusted-signer-named
type: req
---
When the operator refuses an image signed only by signers it doesn't trust, the Plugin's status must name each signer's identity, distinct from an image with no valid signature.

## Context

The author decided it on 2026-11-01, among the rules for the first alpha: until then the cosign adapter kept only signatures that verified as a trusted signer's, so an image signed by anyone else was reported as Unsigned, like an image nobody signed, and the owner couldn't tell a signer they forgot to trust (D-45) from a forgery. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the operator refuses an image signed by a signer it doesn't trust, the `Plugin`'s status names that signer's identity, distinct from an unsigned image, so the owner can trust the signer (D-45) or recognise a forgery.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-169#1
