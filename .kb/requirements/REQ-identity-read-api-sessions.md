---
title: The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token.
logic_claims:
  - CLAIM-C4DB275FE02F68B5
semantic_clauses:
  - The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5762838181d78745223f39643ab958c02408f25638b515484af797fb7b8f8efa
semantic_inventory:
  - claim_key: CLAIM-C4DB275FE02F68B5
    claim_text: The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token
    role: normative
    status: modeled
    span:
      start: 0
      end: 134
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:51.812Z'
id: REQ-identity-read-api-sessions
type: req
---
The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token.

## Context

Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. Author decision, demo read-only sessions (given in chat to the orchestrator, not yet in DESIGN.md): every session carries a scope, read or owner; read may only call read paths (the read API, WhoAmI, Logout); owner adds everything else; the scope is enforced in the core, not in the UI. A demo authenticator issues a read-scoped session without credentials only when the database carries the demo marker AND the core runs with --demo-sessions; it stays single-owner: a demo session is the owner with read scope.

## Source

> Split AUTHENTICATION from the SESSION. ... the core issues an OPAQUE, revocable SESSION token: random secret, stored hashed (SHA-256) like script tokens, with idle and absolute expiry ... not a JWT.
