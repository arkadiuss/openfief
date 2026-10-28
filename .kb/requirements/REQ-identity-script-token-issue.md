---
title: Issuing a script token must create a read-scoped token with a valid name and a lifetime above zero and at most one year, from random bytes of crypto/rand, return its secret only once, and store only the SHA-256 hash of the secret's key.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: Issuing a script token must create a read-scoped token with a valid name and a lifetime above zero and at most one year, from random bytes of crypto/rand, return its secret only once, and store only the SHA-256 hash of the secret's key.
logic_claims:
  - CLAIM-14EF7DB3875CDB23
semantic_clauses:
  - Issuing a script token must create a read-scoped token with a valid name and a lifetime above zero and at most one year, from random bytes of crypto/rand, return its secret only once, and store only the SHA-256 hash of the secret's key.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3203f19665e81502fb2a7cba1707f661503bd251fbf8cd640c49eadb8faa980e
semantic_inventory:
  - claim_key: CLAIM-14EF7DB3875CDB23
    claim_text: Issuing a script token must create a read-scoped token with a valid name and a lifetime above zero and at most one year, from random bytes of crypto/rand, return its secret only once, and store only the SHA-256 hash of the secret's key
    role: normative
    status: modeled
    span:
      start: 0
      end: 235
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:21.873Z'
id: REQ-identity-script-token-issue
type: req
---
Issuing a script token must create a read-scoped token with a valid name and a lifetime above zero and at most one year, from random bytes of crypto/rand, return its secret only once, and store only the SHA-256 hash of the secret's key.

## Context

D-27 gives scripts scoped, expiring, read-only personal tokens; D-29 forbids home-grown cryptography, so the secret comes from crypto/rand and only a SHA-256 hash is kept, which makes a database dump useless against the API. The name and lifetime limits keep listings printable and tokens short-lived. Modeled by Claude Code for the core task: the author asked for minimal script tokens so the read API can run on the demo; the author has not reviewed this requirement yet.

## Source

> Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md D-27

> No home-grown cryptography in M0.

docs/DESIGN.md D-29
