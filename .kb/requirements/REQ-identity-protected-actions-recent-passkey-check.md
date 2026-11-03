---
title: Enrolling or removing a passkey, issuing an enrollment link and creating a script token must need a passkey check made in the same session within the last five minutes, not only an owner session.
status: open
text_ref: design-decisions:D-183#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-183-1
  - lane:strict
semantic_text: Enrolling or removing a passkey, issuing an enrollment link and creating a script token must need a passkey check made in the same session within the last five minutes, not only an owner session.
logic_claims:
  - CLAIM-E81DC28A7B86734E
semantic_clauses:
  - Enrolling or removing a passkey, issuing an enrollment link and creating a script token must need a passkey check made in the same session within the last five minutes, not only an owner session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dd654e7a8e37c3e61f7dc47b36e711c8fefdbe0876265718b9d46d5c15d05cd3
semantic_inventory:
  - claim_key: CLAIM-E81DC28A7B86734E
    claim_text: Enrolling or removing a passkey, issuing an enrollment link and creating a script token must need a passkey check made in the same session within the last five minutes, not only an owner session
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-183#1
  recorded_at: '2026-11-03T02:41:25.892Z'
id: REQ-identity-protected-actions-recent-passkey-check
type: req
---
Enrolling or removing a passkey, issuing an enrollment link and creating a script token must need a passkey check made in the same session within the last five minutes, not only an owner session.

## Context

The author decided it on 2026-11-02, after the security review before the alpha: a stolen session alone must not be able to add a way in, a passkey, an enrollment link or a script token, nor remove the owner's passkeys. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Enrolling or removing a passkey, issuing an enrollment link and creating a script token need a passkey check made in the same session within the last five minutes, not only an owner session, so a stolen session can't add a way in. The owner can sign out every session at once.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-183#1, decided by the author on 2026-11-02
