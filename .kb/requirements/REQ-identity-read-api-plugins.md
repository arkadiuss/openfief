---
title: The read API must accept a bearer token that is no session or script token when the AuthService's caller check accepts it as a view plugin's projected ServiceAccount token with audience openfief-core, must refuse it as an invalid token when that check refuses it, and must fail as internal when that check cannot be made.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: The read API must accept a bearer token that is no session or script token when the AuthService's caller check accepts it as a view plugin's projected ServiceAccount token with audience openfief-core, must refuse it as an invalid token when that check refuses it, and must fail as internal when that check cannot be made.
logic_claims:
  - CLAIM-EA4B749449BAFB02
semantic_clauses:
  - The read API must accept a bearer token that is no session or script token when the AuthService's caller check accepts it as a view plugin's projected ServiceAccount token with audience openfief-core, must refuse it as an invalid token when that check refuses it, and must fail as internal when that check cannot be made.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3c823b05543fe83f75c1f8bbaea1c9075a456bd14fc591a4b1c1c73e8fc3640d
semantic_inventory:
  - claim_key: CLAIM-EA4B749449BAFB02
    claim_text: The read API must accept a bearer token that is no session or script token when the AuthService's caller check accepts it as a view plugin's projected ServiceAccount token with audience openfief-core, must refuse it as an invalid token when that check refuses it, and must fail as internal when that check cannot be made
    role: normative
    status: modeled
    span:
      start: 0
      end: 320
origin:
  kind: agent
  recorded_at: '2026-10-29T01:36:20.189Z'
id: REQ-identity-read-api-plugins
type: req
---
The read API must accept a bearer token that is no session or script token when the AuthService's caller check accepts it as a view plugin's projected ServiceAccount token with audience openfief-core, must refuse it as an invalid token when that check refuses it, and must fail as internal when that check cannot be made.

## Context

D-27: plugins authenticate to the core with projected ServiceAccount tokens; D-6: every view plugin uses the read API. The Ghostfolio view (D-4) syncs on its own schedule with no owner session at hand, so it reads with its own token. The narrowest reading: only view plugins, the callers the AuthService already accepts (TokenReview for audience openfief-core, a Plugin with the view capability), and only the read API, which only reads. Refines REQ-AUTO-C6544ADAF8F5AE65. Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`). Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md D-27

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6
