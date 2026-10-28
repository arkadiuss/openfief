---
title: 'The built-in UI must sign the owner in only through the core''s AuthService: a passkey login and passkey registration relaying the WebAuthn Level 3 JSON untouched, View demo only when the core offers demo sessions, and logout revoking the session.'
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: 'The built-in UI must sign the owner in only through the core''s AuthService: a passkey login and passkey registration relaying the WebAuthn Level 3 JSON untouched, View demo only when the core offers demo sessions, and logout revoking the session.'
logic_claims:
  - CLAIM-B2CE8B3F09E4F95E
semantic_clauses:
  - 'The built-in UI must sign the owner in only through the core''s AuthService: a passkey login and passkey registration relaying the WebAuthn Level 3 JSON untouched, View demo only when the core offers demo sessions, and logout revoking the session.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fa32dcfc79e70dabc8d0a05532fd010c7c15a55457b3cbcb787bda03140762ab
semantic_inventory:
  - claim_key: CLAIM-B2CE8B3F09E4F95E
    claim_text: 'The built-in UI must sign the owner in only through the core''s AuthService: a passkey login and passkey registration relaying the WebAuthn Level 3 JSON untouched, View demo only when the core offers demo sessions, and logout revoking the session'
    role: normative
    status: modeled
    span:
      start: 0
      end: 245
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:28.949Z'
id: REQ-ui-login-flows
type: req
---
The built-in UI must sign the owner in only through the core's AuthService: a passkey login and passkey registration relaying the WebAuthn Level 3 JSON untouched, View demo only when the core offers demo sessions, and logout revoking the session.

## Context

D-26 (passkeys, no external IdP) and D-60 (first passkey through a one-time link). The core's AuthService contract (proto/openfief/identity/v1/auth.proto) and the author's auth split (authenticators, then an opaque revocable session with a read or owner scope) given in chat to the orchestrator.

## Source

> People log in with passkeys (WebAuthn), built in. No external identity provider.

docs/DESIGN.md#D-26
