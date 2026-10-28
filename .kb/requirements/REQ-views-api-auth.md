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

D-27 gives scripts scoped, read-only personal tokens and plugins projected ServiceAccount tokens; both reach the API as bearer tokens. Verifying them belongs to the identity context, which comes later, so the API checks through a consumer-side TokenVerifier port in views/app. Even behind kubectl port-forward (D-28) a user's own Ingress can expose the API, so it never answers unauthenticated. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`). Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md D-27
