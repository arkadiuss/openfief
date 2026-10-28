---
title: After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT.
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT.
logic_claims:
  - CLAIM-680DD6D38E718879
semantic_clauses:
  - After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0f2417960789588a38980617e322b267920e7fc15f7bd3fed6cf02e6922427c4
semantic_inventory:
  - claim_key: CLAIM-680DD6D38E718879
    claim_text: After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT
    role: normative
    status: modeled
    span:
      start: 0
      end: 248
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:09.820Z'
id: REQ-identity-session-opaque
type: req
---
After an authenticator proves the owner, the core must issue an opaque, revocable session token stored only as the SHA-256 hash of its random key, with an idle expiry (12 hours by default) and an absolute expiry (7 days by default), and never a JWT.

## Context

Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. D-29: standard primitives only.

## Source

> Split AUTHENTICATION from the SESSION. ... the core issues an OPAQUE, revocable SESSION token: random secret, stored hashed (SHA-256) like script tokens, with idle and absolute expiry ... not a JWT.
