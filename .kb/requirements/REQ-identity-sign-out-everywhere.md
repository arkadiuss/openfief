---
title: The owner must be able to sign out every session at once.
status: open
text_ref: design-decisions:D-183#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-183-2
  - lane:strict
semantic_text: The owner must be able to sign out every session at once.
logic_claims:
  - CLAIM-C050D6B85246D67D
semantic_clauses:
  - The owner must be able to sign out every session at once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d59498317ba4ff0303794934cd9ea4f3bfd5ef84c475c4060e5ebd1ed57f92da
semantic_inventory:
  - claim_key: CLAIM-C050D6B85246D67D
    claim_text: The owner must be able to sign out every session at once
    role: normative
    status: modeled
    span:
      start: 0
      end: 56
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-183#2
  recorded_at: '2026-11-03T02:41:35.264Z'
id: REQ-identity-sign-out-everywhere
type: req
---
The owner must be able to sign out every session at once.

## Context

The author decided it on 2026-11-02 with the passkey check, after the security review before the alpha: when a device is lost or a session may be stolen, the owner ends every session in one step. D-66 makes sessions opaque and revocable. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Enrolling or removing a passkey, issuing an enrollment link and creating a script token need a passkey check made in the same session within the last five minutes, not only an owner session, so a stolen session can't add a way in. The owner can sign out every session at once.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-183#2, decided by the author on 2026-11-02
