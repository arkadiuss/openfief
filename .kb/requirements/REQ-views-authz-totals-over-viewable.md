---
title: Totals, the allocation and performance must be computed over the accounts the caller may view.
status: open
text_ref: design-decisions:D-163#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-3
  - lane:strict
semantic_text: Totals, the allocation and performance must be computed over the accounts the caller may view.
logic_claims:
  - CLAIM-8B2354CE4495E70C
semantic_clauses:
  - Totals, the allocation and performance must be computed over the accounts the caller may view.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 95ee9c513b487bba21e41ddf754e024260049a46997ba3835b9e63cf5c64c8d3
semantic_inventory:
  - claim_key: CLAIM-8B2354CE4495E70C
    claim_text: Totals, the allocation and performance must be computed over the accounts the caller may view
    role: normative
    status: modeled
    span:
      start: 0
      end: 93
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-163#3
  recorded_at: '2026-11-02T00:56:56.036Z'
id: REQ-views-authz-totals-over-viewable
type: req
---
Totals, the allocation and performance must be computed over the accounts the caller may view.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. A viewer granted some accounts (D-65) would otherwise see the others through a total. D-106 makes valuation totals include cash, D-30 lists the allocation, and D-61 measures performance as a time-weighted return. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#3
