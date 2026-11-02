---
title: Every read path and owner service of the core must ask one authorization interface whether the caller may act on each account.
status: open
text_ref: design-decisions:D-163#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-2
  - lane:strict
semantic_text: Every read path and owner service of the core must ask one authorization interface whether the caller may act on each account.
logic_claims:
  - CLAIM-1FB97F7702E2E188
semantic_clauses:
  - Every read path and owner service of the core must ask one authorization interface whether the caller may act on each account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 682fca9b92689deea579e04d6bd2728f3a062794cfad5c649319327ca31f73c3
semantic_inventory:
  - claim_key: CLAIM-1FB97F7702E2E188
    claim_text: Every read path and owner service of the core must ask one authorization interface whether the caller may act on each account
    role: normative
    status: modeled
    span:
      start: 0
      end: 125
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-163#2
  recorded_at: '2026-11-02T00:56:50.401Z'
id: REQ-identity-authz-one-interface
type: req
---
Every read path and owner service of the core must ask one authorization interface whether the caller may act on each account.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. The read API (D-50) and the owner services (D-79) are the core's read paths and owner services. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#2
