---
title: When the core refuses an action for want of a recent passkey check, the built-in UI must ask the owner to confirm with their passkey, run the passkey check, and post the same action again.
status: open
text_ref: docs/DESIGN.md#D-183
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-183
  - lane:strict
semantic_text: When the core refuses an action for want of a recent passkey check, the built-in UI must ask the owner to confirm with their passkey, run the passkey check, and post the same action again.
logic_claims:
  - CLAIM-BB5088BE75C860C5
semantic_clauses:
  - When the core refuses an action for want of a recent passkey check, the built-in UI must ask the owner to confirm with their passkey, run the passkey check, and post the same action again.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a65b316fc7c1b23e7eb89e073e5f1331fca8d5db9750d895520eedf19e938e4e
semantic_inventory:
  - claim_key: CLAIM-BB5088BE75C860C5
    claim_text: When the core refuses an action for want of a recent passkey check, the built-in UI must ask the owner to confirm with their passkey, run the passkey check, and post the same action again
    role: condition
    status: modeled
    span:
      start: 0
      end: 187
origin:
  kind: agent
  recorded_at: '2026-11-03T02:42:01.630Z'
id: REQ-ui-passkey-check-prompt
type: req
---
When the core refuses an action for want of a recent passkey check, the built-in UI must ask the owner to confirm with their passkey, run the passkey check, and post the same action again.

## Context

Refines REQ-identity-protected-actions-recent-passkey-check (D-183) for the built-in UI, a backend-for-frontend (D-69): the settings page keeps the refused action's path and fields, and webauthn.js runs the assertion and posts them again; adding a passkey on this host runs the check before its own ceremony. Modeled by Claude Code for the last fixes before the alpha tag, as the orchestrator asked on 2026-11-03.

## Source

> Enrolling or removing a passkey, issuing an enrollment link and creating a script token need a passkey check made in the same session within the last five minutes, not only an owner session, so a stolen session can't add a way in. The owner can sign out every session at once.

docs/DESIGN.md D-183
