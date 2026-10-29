---
title: The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise.
status: open
text_ref: design-decisions:D-68#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-68-1
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
  kind: import
  ref: design-decisions:D-68#1
  recorded_at: '2026-10-29T05:29:15.589Z'
id: REQ-identity-demo-sessions
type: req
firmness: firm
required_evidence: asserted
---
The core must issue a credential-free, read-scoped session for the owner only when it runs with demo sessions turned on and the database carries the demo marker, and refuse otherwise.

## Context

Author decision, demo read-only sessions (given in chat to the orchestrator, since recorded in docs/DESIGN.md as D-66 and D-68): every session carries a scope, read or owner; read may only call read paths (the read API, WhoAmI, Logout); owner adds everything else; the scope is enforced in the core, not in the UI. A demo authenticator issues a read-scoped session without credentials only when the database carries the demo marker AND the core runs with --demo-sessions; it stays single-owner: a demo session is the owner with read scope. D-68 makes this requirement the author's; the demo marker is the marker of the synthetic demo dataset, and demo sessions are turned on with --demo-sessions. REQ-identity-demo-owner-passkey records the rest of D-68.

## Source

> A demo authenticator issues `read` sessions without credentials, only when the database carries the marker of the synthetic demo dataset and the core runs with `--demo-sessions`. On a public demo anyone can browse read-only, and the owner logs in with a passkey for an `owner` session.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-68#1

> Session scopes: every session carries a scope, read or owner. ... the demo authenticator ... issues a read-scoped session without credentials, ONLY when the database carries the demo marker ... AND the core runs with --demo-sessions.

Author decision in chat, demo read-only sessions (before D-68 was recorded)
