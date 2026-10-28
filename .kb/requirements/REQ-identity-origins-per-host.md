---
title: The core must accept the owner's browser only on origins on an explicit allow-list, and a passkey or session must work only on the host it was created or issued on.
status: open
text_ref: docs/DESIGN.md#D-60
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-60
  - lane:strict
semantic_text: The core must accept the owner's browser only on origins on an explicit allow-list, and a passkey or session must work only on the host it was created or issued on.
logic_claims:
  - CLAIM-19E5277AF8DDF149
semantic_clauses:
  - The core must accept the owner's browser only on origins on an explicit allow-list, and a passkey or session must work only on the host it was created or issued on.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ea09f6bf43cb575977d89bdb26cadb9eb6379dc81dfe166bbf1f11a69be2aa1f
semantic_inventory:
  - claim_key: CLAIM-19E5277AF8DDF149
    claim_text: The core must accept the owner's browser only on origins on an explicit allow-list, and a passkey or session must work only on the host it was created or issued on
    role: normative
    status: modeled
    span:
      start: 0
      end: 163
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:40.573Z'
id: REQ-identity-origins-per-host
type: req
---
The core must accept the owner's browser only on origins on an explicit allow-list, and a passkey or session must work only on the host it was created or issued on.

## Context

D-60: passkeys are kept per hostname. Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token.

## Source

> [D-60] OpenFief installs with no hostname. A passkey works only on the hostname it was created on, so the CLI prints a one-time link with which the owner creates the first passkey on localhost ... Passkeys are kept per hostname.
