---
title: The AuthService must serve only callers proven to be view plugins by a projected ServiceAccount token with audience openfief-core.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: The AuthService must serve only callers proven to be view plugins by a projected ServiceAccount token with audience openfief-core.
logic_claims:
  - CLAIM-A1CA2554F28A914D
semantic_clauses:
  - The AuthService must serve only callers proven to be view plugins by a projected ServiceAccount token with audience openfief-core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 75671b5881e24a5c11f41d32b87d734f7eba81c9d338fe623bba825e9ddaf38d
semantic_inventory:
  - claim_key: CLAIM-A1CA2554F28A914D
    claim_text: The AuthService must serve only callers proven to be view plugins by a projected ServiceAccount token with audience openfief-core
    role: normative
    status: modeled
    span:
      start: 0
      end: 129
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:56.445Z'
id: REQ-identity-auth-callers
type: req
---
The AuthService must serve only callers proven to be view plugins by a projected ServiceAccount token with audience openfief-core.

## Context

Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. D-27: plugins authenticate with projected ServiceAccount tokens, audience openfief-core.

## Source

> Split AUTHENTICATION from the SESSION. ... the core issues an OPAQUE, revocable SESSION token: random secret, stored hashed (SHA-256) like script tokens, with idle and absolute expiry ... not a JWT.
