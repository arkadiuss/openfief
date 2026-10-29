---
title: Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass.
logic_claims:
  - CLAIM-5C6A8447F7C348E6
semantic_clauses:
  - Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cce48febf9b4d70b2935b84258597c42191091f94e82f47bfceda4ccb003d6f4
semantic_inventory:
  - claim_key: CLAIM-5C6A8447F7C348E6
    claim_text: Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass
    role: normative
    status: modeled
    span:
      start: 0
      end: 119
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:45.415Z'
id: REQ-identity-script-token-from-session
type: req
---
Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass.

## Context

Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. D-27: scripts get scoped, expiring, read-only personal tokens.

## Source

> Split AUTHENTICATION from the SESSION. ... the core issues an OPAQUE, revocable SESSION token: random secret, stored hashed (SHA-256) like script tokens, with idle and absolute expiry ... not a JWT.
