---
title: The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise.
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise.
logic_claims:
  - CLAIM-4A1854C9675DFB31
semantic_clauses:
  - The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9348bc89c80fffc406883b67f705361aa0e3f63e42ee641371e4c317dec4ecd8
semantic_inventory:
  - claim_key: CLAIM-4A1854C9675DFB31
    claim_text: The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise
    role: normative
    status: modeled
    span:
      start: 0
      end: 182
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:20.915Z'
id: REQ-identity-demo-sessions
type: req
---
The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise.

## Context

Author decision, demo read-only sessions (given in chat to the orchestrator, not yet in DESIGN.md): every session carries a scope, read or owner; read may only call read paths (the read API, WhoAmI, Logout); owner adds everything else; the scope is enforced in the core, not in the UI. A demo authenticator issues a read-scoped session without credentials only when the database carries the demo marker AND the core runs with --demo-sessions; it stays single-owner: a demo session is the owner with read scope.

## Source

> Session scopes: every session carries a scope, read or owner. ... the demo authenticator ... issues a read-scoped session without credentials, ONLY when the database carries the demo marker ... AND the core runs with --demo-sessions.
