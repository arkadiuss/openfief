---
title: 'Every session must carry a scope, read or owner, enforced by the core: a read session may only use the read API, WhoAmI and Logout, while an owner session may also manage passkeys, script tokens and enrollment links.'
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: 'Every session must carry a scope, read or owner, enforced by the core: a read session may only use the read API, WhoAmI and Logout, while an owner session may also manage passkeys, script tokens and enrollment links.'
logic_claims:
  - CLAIM-06A52880D769876F
semantic_clauses:
  - 'Every session must carry a scope, read or owner, enforced by the core: a read session may only use the read API, WhoAmI and Logout, while an owner session may also manage passkeys, script tokens and enrollment links.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b83646ff79f0b6dc1da9c57975db56a1732ad6c134d4297c773b391d8a219ff5
semantic_inventory:
  - claim_key: CLAIM-06A52880D769876F
    claim_text: 'Every session must carry a scope, read or owner, enforced by the core: a read session may only use the read API, WhoAmI and Logout, while an owner session may also manage passkeys, script tokens and enrollment links'
    role: normative
    status: modeled
    span:
      start: 0
      end: 215
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:14.438Z'
id: REQ-identity-session-scopes
type: req
---
Every session must carry a scope, read or owner, enforced by the core: a read session may only use the read API, WhoAmI and Logout, while an owner session may also manage passkeys, script tokens and enrollment links.

## Context

Author decision, demo read-only sessions (given in chat to the orchestrator, not yet in DESIGN.md): every session carries a scope, read or owner; read may only call read paths (the read API, WhoAmI, Logout); owner adds everything else; the scope is enforced in the core, not in the UI. A demo authenticator issues a read-scoped session without credentials only when the database carries the demo marker AND the core runs with --demo-sessions; it stays single-owner: a demo session is the owner with read scope.

## Source

> Session scopes: every session carries a scope, read or owner. ... the demo authenticator ... issues a read-scoped session without credentials, ONLY when the database carries the demo marker ... AND the core runs with --demo-sessions.
