---
title: The owner lists, per data tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default.
status: open
text_ref: design-decisions:D-206#2
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-206#2
  recorded_at: '2026-11-13T11:28:54.080Z'
semantic_text: The owner lists, per data tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default.
semantic_source_field: semantic_text
semantic_source_hash: 083909b510696b00af031d7f55a52f694376414caaa41ee61ddce94c2359f929
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-AA53CA5798E8F6C5
    claim_text: The owner lists, per data tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default
    role: descriptive
    status: missing
    span:
      start: 0
      end: 161
logic_claims:
  - CLAIM-AA53CA5798E8F6C5
semantic_clauses:
  - The owner lists, per data tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default
required_evidence: asserted
id: REQ-models-data-tiers-deny-by-default
type: req
---
The owner lists, per data tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default.

## Context

Every message of a request carries a tier (T0, TF, T1 or T2); data nobody tagged is not sent; a tier with no models listed allows none.

## Source

> The owner lists, per tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default.

Source: OpenFief author decisions (DESIGN.md section 10) - D-206#2