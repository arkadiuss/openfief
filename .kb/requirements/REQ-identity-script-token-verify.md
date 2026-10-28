---
title: Verifying a script token must accept only a secret whose ID is stored, whose key matches the stored hash in constant time, that is used before its expiry, is not revoked and has the requested scope, and must otherwise fail with an invalid-token error that never contains the secret, while a store failure fails with a different error.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: Verifying a script token must accept only a secret whose ID is stored, whose key matches the stored hash in constant time, that is used before its expiry, is not revoked and has the requested scope, and must otherwise fail with an invalid-token error that never contains the secret, while a store failure fails with a different error.
logic_claims:
  - CLAIM-DC3E19EA220B7EAC
semantic_clauses:
  - Verifying a script token must accept only a secret whose ID is stored, whose key matches the stored hash in constant time, that is used before its expiry, is not revoked and has the requested scope, and must otherwise fail with an invalid-token error that never contains the secret, while a store failure fails with a different error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 49b6b0816ca13b0d9a46629d44931971a860477760616e717378ff5dcb7086ea
semantic_inventory:
  - claim_key: CLAIM-DC3E19EA220B7EAC
    claim_text: Verifying a script token must accept only a secret whose ID is stored, whose key matches the stored hash in constant time, that is used before its expiry, is not revoked and has the requested scope, and must otherwise fail with an invalid-token error that never contains the secret, while a store failure fails with a different error
    role: normative
    status: modeled
    span:
      start: 0
      end: 333
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:26.510Z'
id: REQ-identity-script-token-verify
type: req
---
Verifying a script token must accept only a secret whose ID is stored, whose key matches the stored hash in constant time, that is used before its expiry, is not revoked and has the requested scope, and must otherwise fail with an invalid-token error that never contains the secret, while a store failure fails with a different error.

## Context

The read API asks the identity context whether a bearer token may read (D-27); a constant-time comparison (crypto/subtle, D-29) avoids timing leaks, and keeping store failures apart from invalid tokens lets the API answer 500 instead of 401 when the database is down. Modeled by Claude Code for the core task: the author asked for minimal script tokens so the read API can run on the demo; the author has not reviewed this requirement yet.

## Source

> Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md D-27
