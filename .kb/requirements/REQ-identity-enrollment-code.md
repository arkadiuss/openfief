---
title: Creating a passkey without an owner session must need a one-time enrollment code issued for that origin, stored hashed, that expires (15 minutes by default, 24 hours at most) and works once.
status: open
text_ref: docs/DESIGN.md#D-60
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-60
  - lane:strict
semantic_text: Creating a passkey without an owner session must need a one-time enrollment code issued for that origin, stored hashed, that expires (15 minutes by default, 24 hours at most) and works once.
logic_claims:
  - CLAIM-63AC0452CD1ED8E2
semantic_clauses:
  - Creating a passkey without an owner session must need a one-time enrollment code issued for that origin, stored hashed, that expires (15 minutes by default, 24 hours at most) and works once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a24a5578f575f9317f2ce552de08a8e80890a14cfd787bdec1115c25719500d5
semantic_inventory:
  - claim_key: CLAIM-63AC0452CD1ED8E2
    claim_text: Creating a passkey without an owner session must need a one-time enrollment code issued for that origin, stored hashed, that expires (15 minutes by default, 24 hours at most) and works once
    role: normative
    status: modeled
    span:
      start: 0
      end: 189
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:34.076Z'
id: REQ-identity-enrollment-code
type: req
---
Creating a passkey without an owner session must need a one-time enrollment code issued for that origin, stored hashed, that expires (15 minutes by default, 24 hours at most) and works once.

## Context

D-60: the CLI prints a one-time link for the first passkey on localhost; a logged-in session issues one for a new hostname. Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token.

## Source

> [D-60] OpenFief installs with no hostname. A passkey works only on the hostname it was created on, so the CLI prints a one-time link with which the owner creates the first passkey on localhost ... Passkeys are kept per hostname.
