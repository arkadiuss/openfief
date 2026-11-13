---
title: The owner revokes a viewer, which ends their sessions, disables their passkeys and ends their grants, each with an audit entry, removes the grants from the authorization model first, and keeps the person record with their ID.
status: open
firmness: default
required_evidence: unverified
semantic_text: The owner revokes a viewer, which ends their sessions, disables their passkeys and ends their grants, each with an audit entry, removes the grants from the authorization model first, and keeps the person record with their ID.
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7180a3d6f9f70e27ecdb0c18cc85956bc96aba257468cae13374ab8a2b18f71d
semantic_inventory:
  - claim_key: CLAIM-64ED295DA09F963C
    claim_text: The owner revokes a viewer, which ends their sessions, disables their passkeys and ends their grants, each with an audit entry, removes the grants from the authorization model first, and keeps the person record with their ID
    role: descriptive
    status: missing
    span:
      start: 0
      end: 224
logic_claims:
  - CLAIM-64ED295DA09F963C
id: REQ-identity-revoke-viewer
type: req
---
The owner revokes a viewer, which ends their sessions, disables their passkeys and ends their grants, each with an audit entry, removes the grants from the authorization model first, and keeps the person record with their ID.

## Context

Q8: persons are never hard-deleted. Revoking only takes access away, so it needs no passkey check; the owner cannot be revoked, and a viewer revoked already is not found. Decisions: D-210, D-211, Q8.

## Source

> S6 task brief and the permissions plan's author and design-session answers (Q6, Q7, Q8), 2026-11-13.