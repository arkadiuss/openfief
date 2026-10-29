---
title: The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token.
status: open
text_ref: design-decisions:D-83#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-83-1
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
  kind: import
  ref: design-decisions:D-83#1
  recorded_at: '2026-10-29T08:19:39.750Z'
id: REQ-identity-read-api-sessions
type: req
firmness: firm
required_evidence: asserted
---
The read API must authenticate a bearer token as the owner when it is a valid session of any scope or a valid read-scoped script token.

## Context

Author decision, auth split (given in chat to the orchestrator, since recorded in docs/DESIGN.md as D-65 to D-69): authentication is split from the session; authenticators (passkeys now) all resolve to the single owner (GR-2); after authentication the core issues an opaque, revocable session token, random, stored as a SHA-256 hash, with idle and absolute expiry, not a JWT; script tokens are created only by a logged-in session, and core token issue stays as break-glass; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. Author decision, demo read-only sessions (given in chat to the orchestrator, since recorded as D-66 and D-68): every session carries a scope, read or owner; read may only call read paths (the read API, WhoAmI, Logout); owner adds everything else; the scope is enforced in the core, not in the UI. A demo authenticator issues a read-scoped session without credentials only when the database carries the demo marker AND the core runs with --demo-sessions; it stays single-owner: a demo session is the owner with read scope. D-83 makes this requirement the author's: the read API accepts sessions of either scope and script tokens, which D-27 makes read-only. REQ-identity-read-api-view-plugin-tokens and REQ-identity-read-api-refuses-connector-tokens record the rest of D-83.

## Source

> The read API accepts sessions of either scope, script tokens, and view plugins' own tokens. It refuses connector tokens: the core calls connectors (D-49), and a connector never needs to read the owner's record.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-83#1

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`). Scripts get separate, scoped, expiring, read-only personal tokens.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-27

> Split AUTHENTICATION from the SESSION. ... the core issues an OPAQUE, revocable SESSION token: random secret, stored hashed (SHA-256) like script tokens, with idle and absolute expiry ... not a JWT.

Author decision in chat, auth split (before D-65 to D-69 were recorded)
