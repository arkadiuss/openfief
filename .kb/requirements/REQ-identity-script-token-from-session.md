---
title: Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass.
status: open
text_ref: design-decisions:D-67#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-67-1
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
  kind: import
  ref: design-decisions:D-67#1
  recorded_at: '2026-10-29T05:28:14.239Z'
id: REQ-identity-script-token-from-session
type: req
firmness: firm
required_evidence: asserted
---
Script tokens must be created only from an owner session, with core token issue kept as the administrator's break-glass.

## Context

Author decision, auth split (given in chat to the orchestrator, since recorded in docs/DESIGN.md as D-65 to D-69): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. D-27: scripts get scoped, expiring, read-only personal tokens. D-67 makes this requirement the author's; REQ-identity-no-preinstalled-token records the rest of D-67.

## Source

> Script tokens (D-27) are created only from an owner session in the UI, and no token is installed in advance as a Secret. `core token issue`, run in the core's pod, stays as a break-glass for an administrator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-67#1

> Split AUTHENTICATION from the SESSION. ... the core issues an OPAQUE, revocable SESSION token: random secret, stored hashed (SHA-256) like script tokens, with idle and absolute expiry ... not a JWT.

Author decision in chat, auth split (before D-67 was recorded)
