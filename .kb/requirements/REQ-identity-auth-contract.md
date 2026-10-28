---
title: The core must offer view plugins an authentication service, openfief.identity.v1.AuthService over Connect RPC, separate from the read-only public API, for passkey registration and login, demo read-only sessions, sessions with a read or owner scope, script tokens and enrollment links.
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: The core must offer view plugins an authentication service, openfief.identity.v1.AuthService over Connect RPC, separate from the read-only public API, for passkey registration and login, demo read-only sessions, sessions with a read or owner scope, script tokens and enrollment links.
logic_claims:
  - CLAIM-DCF87CA2A55E5311
semantic_clauses:
  - The core must offer view plugins an authentication service, openfief.identity.v1.AuthService over Connect RPC, separate from the read-only public API, for passkey registration and login, demo read-only sessions, sessions with a read or owner scope, script tokens and enrollment links.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4da4f2384bf9d3b57ffe308ddd2a28c3c9a1a1ac651f4d78ae35d6589d35629b
semantic_inventory:
  - claim_key: CLAIM-DCF87CA2A55E5311
    claim_text: The core must offer view plugins an authentication service, openfief.identity.v1.AuthService over Connect RPC, separate from the read-only public API, for passkey registration and login, demo read-only sessions, sessions with a read or owner scope, script tokens and enrollment links
    role: normative
    status: modeled
    span:
      start: 0
      end: 283
origin:
  kind: agent
  recorded_at: '2026-10-28T18:35:01.901Z'
id: REQ-identity-auth-contract
type: req
---
The core must offer view plugins an authentication service, openfief.identity.v1.AuthService over Connect RPC, separate from the read-only public API, for passkey registration and login, demo read-only sessions, sessions with a read or owner scope, script tokens and enrollment links.

## Context

Author decision, auth split (given in chat to the orchestrator, not yet in DESIGN.md): authentication is split from the session; the built-in UI is a view plugin acting as backend-for-frontend and calls the core's auth service with its own projected ServiceAccount token. Author decision, demo read-only sessions: every session carries a read or owner scope, and a demo authenticator issues read sessions on demo databases. D-26 (passkeys), D-6 (the public API is read-only), D-11 (protobuf contracts over Connect RPC).

## Source

> [D-26] People log in with passkeys (WebAuthn), built in. No external identity provider.

> [D-6] OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned /v1.

docs/DESIGN.md#D-26; author decisions in chat: auth split; demo read-only sessions
