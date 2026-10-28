---
title: Passkey registration and login must run through the go-webauthn library with discoverable credentials and user verification required, and a login whose signature counter did not increase must be refused.
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: Passkey registration and login must run through the go-webauthn library with discoverable credentials and user verification required, and a login whose signature counter did not increase must be refused.
logic_claims:
  - CLAIM-F6D6C01F4DA9E636
semantic_clauses:
  - Passkey registration and login must run through the go-webauthn library with discoverable credentials and user verification required, and a login whose signature counter did not increase must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c60414a8a5f3e822570df8b94761923778639e97caf6f92bbe2bd27afcb56690
semantic_inventory:
  - claim_key: CLAIM-F6D6C01F4DA9E636
    claim_text: Passkey registration and login must run through the go-webauthn library with discoverable credentials and user verification required, and a login whose signature counter did not increase must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:27.503Z'
id: REQ-identity-passkey-ceremonies
type: req
---
Passkey registration and login must run through the go-webauthn library with discoverable credentials and user verification required, and a login whose signature counter did not increase must be refused.

## Context

Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. D-26 (passkeys), D-29 (no home-grown cryptography).

## Source

> [D-26] People log in with passkeys (WebAuthn), built in. No external identity provider.
