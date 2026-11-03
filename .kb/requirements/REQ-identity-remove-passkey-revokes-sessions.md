---
title: Removing a passkey must revoke every session that passkey signed in, in the same transaction.
status: open
text_ref: docs/DESIGN.md#D-183
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-183
  - lane:strict
semantic_text: Removing a passkey must revoke every session that passkey signed in, in the same transaction.
logic_claims:
  - CLAIM-6F599805071778E7
semantic_clauses:
  - Removing a passkey must revoke every session that passkey signed in, in the same transaction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b97c59997ff8690dd51525baff5afd3afb79630071f58c46d9cd7db84890a1c5
semantic_inventory:
  - claim_key: CLAIM-6F599805071778E7
    claim_text: Removing a passkey must revoke every session that passkey signed in, in the same transaction
    role: normative
    status: modeled
    span:
      start: 0
      end: 92
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:54.853Z'
id: REQ-identity-remove-passkey-revokes-sessions
type: req
---
Removing a passkey must revoke every session that passkey signed in, in the same transaction.

## Context

Refines D-183: a passkey is removed because it may be lost or stolen, so the sessions it opened end with it. Each session records the credential ID of the passkey that signed it in; a later passkey check with another passkey doesn't change it. The orchestrator asked for this reading on 2026-11-03. Modeled by Claude Code for the last fixes before the alpha tag, as the orchestrator asked on 2026-11-03.

## Source

> Enrolling or removing a passkey, issuing an enrollment link and creating a script token need a passkey check made in the same session within the last five minutes, not only an owner session, so a stolen session can't add a way in. The owner can sign out every session at once.

docs/DESIGN.md D-183
