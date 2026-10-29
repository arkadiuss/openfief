---
title: 'No build of OpenFief may accept a fixed token or any other way around authentication: outside a cluster the core must refuse every AuthService caller, and the built-in UI must present only its projected ServiceAccount token.'
status: open
text_ref: docs/DESIGN.md#D-84
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-84
  - lane:strict
semantic_text: 'No build of OpenFief may accept a fixed token or any other way around authentication: outside a cluster the core must refuse every AuthService caller, and the built-in UI must present only its projected ServiceAccount token.'
logic_claims:
  - CLAIM-470CEEBB591CE491
semantic_clauses:
  - 'No build of OpenFief may accept a fixed token or any other way around authentication: outside a cluster the core must refuse every AuthService caller, and the built-in UI must present only its projected ServiceAccount token.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a36334f504ad98881abf42257c2cde679a84273a1c6d852537e82a1df120c506
semantic_inventory:
  - claim_key: CLAIM-470CEEBB591CE491
    claim_text: 'No build of OpenFief may accept a fixed token or any other way around authentication: outside a cluster the core must refuse every AuthService caller, and the built-in UI must present only its projected ServiceAccount token'
    role: normative
    status: modeled
    span:
      start: 0
      end: 223
origin:
  kind: agent
  recorded_at: '2026-10-29T10:54:23.857Z'
id: REQ-identity-no-auth-bypass
type: req
---
No build of OpenFief may accept a fixed token or any other way around authentication: outside a cluster the core must refuse every AuthService caller, and the built-in UI must present only its projected ServiceAccount token.

## Context

Author decision D-84 (docs/DESIGN.md): no fixed token or other way around authentication, not even for development; local UI development runs against kind or k3d. The core's OPENFIEF_AUTH_CALLER_TOKEN path and the UI's matching variable were removed; tests admit the view plugin through a fake cluster's TokenReview instead.

## Source

> [D-84] No build of OpenFief has a fixed token or any other way around authentication, not even for development. Local development of the UI runs against a local cluster such as kind or k3d.

docs/DESIGN.md#D-84
