---
title: Until SpiceDB answers the authorization interface, its only implementation must let the owner act on every account.
status: open
text_ref: design-decisions:D-163#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-4
  - lane:strict
semantic_text: Until SpiceDB answers the authorization interface, its only implementation must let the owner act on every account.
logic_claims:
  - CLAIM-D9719380F847E2D1
semantic_clauses:
  - Until SpiceDB answers the authorization interface, its only implementation must let the owner act on every account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: deff1df67530ce5883b01caea434672d1cf5821b42f811946425a13991a776c3
semantic_inventory:
  - claim_key: CLAIM-D9719380F847E2D1
    claim_text: Until SpiceDB answers the authorization interface, its only implementation must let the owner act on every account
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-163#4
  recorded_at: '2026-11-01T22:45:15.062Z'
id: REQ-identity-authz-owner-acts-every-account
type: req
---
Until SpiceDB answers the authorization interface, its only implementation must let the owner act on every account.

## Context

The author decided it on 2026-11-01, when planning viewers for the second alpha (D-65). D-65 has the owner hold every account and alone change anything. identityapp.Policy, over the pure identity.MayAct, is that implementation. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#4
