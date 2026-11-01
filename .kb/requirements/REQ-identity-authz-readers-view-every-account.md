---
title: Until SpiceDB answers the authorization interface, its only implementation must give every other holder of read access a view of every account.
status: open
text_ref: design-decisions:D-163#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-5
  - lane:strict
semantic_text: Until SpiceDB answers the authorization interface, its only implementation must give every other holder of read access a view of every account.
logic_claims:
  - CLAIM-459805860E6D0F12
semantic_clauses:
  - Until SpiceDB answers the authorization interface, its only implementation must give every other holder of read access a view of every account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8114bf356b3007ec3790d0f810aa97e2521945018de0ff8b5fddb5545c3c25a3
semantic_inventory:
  - claim_key: CLAIM-459805860E6D0F12
    claim_text: Until SpiceDB answers the authorization interface, its only implementation must give every other holder of read access a view of every account
    role: normative
    status: modeled
    span:
      start: 0
      end: 142
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-163#5
  recorded_at: '2026-11-01T22:45:20.974Z'
id: REQ-identity-authz-readers-view-every-account
type: req
---
Until SpiceDB answers the authorization interface, its only implementation must give every other holder of read access a view of every account.

## Context

The author decided it on 2026-11-01, when planning viewers for the second alpha (D-65). The other holders of read access are script tokens, view plugins and demo sessions (D-137); none of them changes anything. Per-account grants for them come with SpiceDB (D-164). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#5
