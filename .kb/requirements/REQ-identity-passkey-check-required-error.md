---
title: An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again.
status: open
text_ref: docs/DESIGN.md#D-183
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-183
  - lane:strict
semantic_text: An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again.
logic_claims:
  - CLAIM-C484F2B18928522F
semantic_clauses:
  - An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0ebffc7610ce7ce7d8370a986e2320c7b1fab5c8feaecc967977d5adf88bb881
semantic_inventory:
  - claim_key: CLAIM-C484F2B18928522F
    claim_text: An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:45.161Z'
id: REQ-identity-passkey-check-required-error
type: req
---
An action refused for want of a recent passkey check must fail with FAILED_PRECONDITION and a PasskeyCheckRequired error detail naming the five-minute window, so a view plugin can ask for the check and try again.

## Context

Refines REQ-identity-protected-actions-recent-passkey-check (D-183): the built-in UI tells this refusal apart from other failed preconditions by its detail, never by its message. Signing in with a passkey counts as a check, and BeginPasskeyCheck and FinishPasskeyCheck make a new one in the session. Modeled by Claude Code for the last fixes before the alpha tag, as the orchestrator asked on 2026-11-03.

## Source

> Enrolling or removing a passkey, issuing an enrollment link and creating a script token need a passkey check made in the same session within the last five minutes, not only an owner session, so a stolen session can't add a way in. The owner can sign out every session at once.

docs/DESIGN.md D-183
