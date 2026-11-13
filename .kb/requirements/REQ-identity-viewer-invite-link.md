---
title: The owner issues a viewer a one-time, expiring enrollment link, in an owner session with a recent passkey check, that carries the name the owner gave them and the grants they start with, each of which must exist.
status: open
firmness: default
required_evidence: unverified
semantic_text: The owner issues a viewer a one-time, expiring enrollment link, in an owner session with a recent passkey check, that carries the name the owner gave them and the grants they start with, each of which must exist.
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2ac185836e20a14bf525578effe991488bfce92edcb7a58d1eba509d724fa82e
semantic_inventory:
  - claim_key: CLAIM-906E27B9D242D976
    claim_text: The owner issues a viewer a one-time, expiring enrollment link, in an owner session with a recent passkey check, that carries the name the owner gave them and the grants they start with, each of which must exist
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 211
logic_claims:
  - CLAIM-906E27B9D242D976
id: REQ-identity-viewer-invite-link
type: req
---
The owner issues a viewer a one-time, expiring enrollment link, in an owner session with a recent passkey check, that carries the name the owner gave them and the grants they start with, each of which must exist.

## Context

InviteViewer on the AuthService. The link is the owner's enrollment code with a name and grants added, so it works once, on one origin, within its lifetime; nothing is made until it is used. A viewer's session, a demo session or an owner session without a recent passkey check is refused. Decisions: D-65, D-210.

## Source

> S6 task brief and the permissions plan's author and design-session answers (Q6, Q7, Q8), 2026-11-13.