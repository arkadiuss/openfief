---
title: Every read API call, over any protocol, must carry a bearer token that the token verifier accepts; otherwise it must be refused as unauthenticated with a Bearer challenge and no data, and a token that cannot be checked must fail as internal.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: Every read API call, over any protocol, must carry a bearer token that the token verifier accepts; otherwise it must be refused as unauthenticated with a Bearer challenge and no data, and a token that cannot be checked must fail as internal.
logic_claims:
  - CLAIM-AE6725896A2F811B
semantic_clauses:
  - Every read API call, over any protocol, must carry a bearer token that the token verifier accepts; otherwise it must be refused as unauthenticated with a Bearer challenge and no data, and a token that cannot be checked must fail as internal.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8bfc607a5731c5af0e44f5ecf05b9f54cf961d38521025a57413b0144bbaf335
semantic_inventory:
  - claim_key: CLAIM-AE6725896A2F811B
    claim_text: Every read API call, over any protocol, must carry a bearer token that the token verifier accepts; otherwise it must be refused as unauthenticated with a Bearer challenge and no data, and a token that cannot be checked must fail as internal
    role: normative
    status: modeled
    span:
      start: 0
      end: 240
origin:
  kind: agent
  recorded_at: '2026-10-28T08:44:16.253Z'
id: REQ-views-api-auth
type: req
---
Every read API call, over any protocol, must carry a bearer token that the token verifier accepts; otherwise it must be refused as unauthenticated with a Bearer challenge and no data, and a token that cannot be checked must fail as internal.

## Context

D-27 gives scripts read-only tokens and D-83 lists the read API's callers. The read API is on both of the core's ports (D-115): the browser port, which the owner's route or a port-forward reaches, and the plugin port, which view plugins call; every call on either needs a token.

## Source

> Scripts and view plugins read through the API with tokens (D-27); the read API accepts sessions of either scope, script tokens, and view plugins' own tokens (D-83).

docs/DESIGN.md#D-83
