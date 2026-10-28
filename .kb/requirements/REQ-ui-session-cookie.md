---
title: The built-in UI must keep the owner's session token only in a __Host- cookie that is HttpOnly, Secure and SameSite=Strict with Path=/, and must send the browser to the login page and delete the cookie when the core no longer accepts the session.
status: open
text_ref: docs/DESIGN.md#D-26
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-26
  - lane:strict
semantic_text: The built-in UI must keep the owner's session token only in a __Host- cookie that is HttpOnly, Secure and SameSite=Strict with Path=/, and must send the browser to the login page and delete the cookie when the core no longer accepts the session.
logic_claims:
  - CLAIM-A7102A06A9E2D08D
semantic_clauses:
  - The built-in UI must keep the owner's session token only in a __Host- cookie that is HttpOnly, Secure and SameSite=Strict with Path=/, and must send the browser to the login page and delete the cookie when the core no longer accepts the session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5af7a93b0824c753feea025cf34c0c55af16cc10ef38028879809d6a65398bbe
semantic_inventory:
  - claim_key: CLAIM-A7102A06A9E2D08D
    claim_text: The built-in UI must keep the owner's session token only in a __Host- cookie that is HttpOnly, Secure and SameSite=Strict with Path=/, and must send the browser to the login page and delete the cookie when the core no longer accepts the session
    role: normative
    status: modeled
    span:
      start: 0
      end: 244
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:40.374Z'
id: REQ-ui-session-cookie
type: req
---
The built-in UI must keep the owner's session token only in a __Host- cookie that is HttpOnly, Secure and SameSite=Strict with Path=/, and must send the browser to the login page and delete the cookie when the core no longer accepts the session.

## Context

The orchestrator's task for the UI (BFF: session token in a __Host- cookie, HttpOnly, Secure, SameSite=Strict, Path=/) and the AuthService contract (UNAUTHENTICATED drops the cookie). Secure also applies on http://localhost: a __Host- cookie must be Secure and gosec refuses insecure cookies. The core's AuthService contract (proto/openfief/identity/v1/auth.proto) and the author's auth split (authenticators, then an opaque revocable session with a read or owner scope) given in chat to the orchestrator.

## Source

> The view plugin keeps it in an HttpOnly, Secure (except on http://localhost), SameSite=Strict cookie

proto/openfief/identity/v1/auth.proto
