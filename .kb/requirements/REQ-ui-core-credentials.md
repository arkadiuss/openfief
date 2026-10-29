---
title: The built-in UI must call the core's AuthService with its own projected ServiceAccount token, read from its file on every call, and the read API with the owner's session token as the bearer, never mixing the two.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: The built-in UI must call the core's AuthService with its own projected ServiceAccount token, read from its file on every call, and the read API with the owner's session token as the bearer, never mixing the two.
logic_claims:
  - CLAIM-3581F50D38DBDE86
semantic_clauses:
  - The built-in UI must call the core's AuthService with its own projected ServiceAccount token, read from its file on every call, and the read API with the owner's session token as the bearer, never mixing the two.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c60e0a738bcf18023a4ac80872dca4d1639873ae47d3e974f4dc0e7ea0ec1a60
semantic_inventory:
  - claim_key: CLAIM-3581F50D38DBDE86
    claim_text: The built-in UI must call the core's AuthService with its own projected ServiceAccount token, read from its file on every call, and the read API with the owner's session token as the bearer, never mixing the two
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:34.697Z'
id: REQ-ui-core-credentials
type: req
---
The built-in UI must call the core's AuthService with its own projected ServiceAccount token, read from its file on every call, and the read API with the owner's session token as the bearer, never mixing the two.

## Context

D-27 (plugins authenticate with projected tokens, audience openfief-core) and the AuthService contract's calling conventions. The core's AuthService contract (proto/openfief/identity/v1/auth.proto) and the author's auth split (authenticators, then an opaque revocable session with a read or owner scope) given in chat to the orchestrator.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`).

docs/DESIGN.md#D-27
