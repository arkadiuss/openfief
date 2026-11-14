---
title: The actor of an audit entry must be whoever asked, the owner or a viewer, by the ID of their session, which the log resolves to the person for ever, and never the owner for a viewer's question; the maintainer's weekly runs are the system's.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:39.496Z'
semantic_text: The actor of an audit entry must be whoever asked, the owner or a viewer, by the ID of their session, which the log resolves to the person for ever, and never the owner for a viewer's question; the maintainer's weekly runs are the system's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3682b55d0eb87ddf5bb313811a7fc1f3a3aaa6b9beac9216b48183b64b3a599d
semantic_inventory:
  - claim_key: CLAIM-61CBB04B50580CD6
    claim_text: The actor of an audit entry must be whoever asked, the owner or a viewer, by the ID of their session, which the log resolves to the person for ever, and never the owner for a viewer's question
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 192
  - claim_key: CLAIM-4649CE504DA3292B
    claim_text: the maintainer's weekly runs are the system's
    role: descriptive
    status: missing
    span:
      start: 194
      end: 239
logic_claims:
  - CLAIM-61CBB04B50580CD6
  - CLAIM-4649CE504DA3292B
id: REQ-identity-audit-viewer-actor
type: req
---
The actor of an audit entry must be whoever asked, the owner or a viewer, by the ID of their session, which the log resolves to the person for ever, and never the owner for a viewer's question; the maintainer's weekly runs are the system's.

## Context

S9 found that the Ask box recorded any session that was not the system as the owner, so a viewer's question was entered as the owner's. Sessions and people are never deleted, so an entry made in a session that ended, or by a person who was revoked, still names the person. The other places that map a session to an actor (the owner services, the maintainer's apply and dismiss, the passkey ceremony) are owner-only or already tell a viewer apart.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14