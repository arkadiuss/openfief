---
title: A viewer's session has read scope, so every call that changes something, the identity service's and every other service's, refuses it with permission denied; a viewer's passkey never confirms the owner's session, and signing out is all a viewer changes.
status: open
firmness: default
required_evidence: unverified
semantic_text: A viewer's session has read scope, so every call that changes something, the identity service's and every other service's, refuses it with permission denied; a viewer's passkey never confirms the owner's session, and signing out is all a viewer changes.
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e4a6d89c8a0abd1e584067e5b7280cd99a47c0090afb93bdc46ab2e0f07df8ad
semantic_inventory:
  - claim_key: CLAIM-8788A865AB0A31D9
    claim_text: A viewer's session has read scope, so every call that changes something, the identity service's and every other service's, refuses it with permission denied
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 156
  - claim_key: CLAIM-462A663AC6A22F04
    claim_text: a viewer's passkey never confirms the owner's session, and signing out is all a viewer changes
    role: descriptive
    status: missing
    span:
      start: 158
      end: 252
logic_claims:
  - CLAIM-8788A865AB0A31D9
  - CLAIM-462A663AC6A22F04
id: REQ-identity-viewer-session-reads-only
type: req
---
A viewer's session has read scope, so every call that changes something, the identity service's and every other service's, refuses it with permission denied; a viewer's passkey never confirms the owner's session, and signing out is all a viewer changes.

## Context

One central check: owner services require an owner-scoped session. The viewer's session is a passkey session with read scope, which a table check also refuses to be owner-scoped. Decisions: D-65, Q6.

## Source

> S6 task brief and the permissions plan's author and design-session answers (Q6, Q7, Q8), 2026-11-13.