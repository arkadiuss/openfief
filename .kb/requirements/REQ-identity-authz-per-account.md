---
title: Authorization must be decided per account and be relationship-based.
status: open
text_ref: design-decisions:D-163#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-163-1
  - lane:strict
semantic_text: Authorization must be decided per account and be relationship-based.
logic_claims:
  - CLAIM-5021B0234869E5FB
semantic_clauses:
  - Authorization must be decided per account and be relationship-based.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b89083bc3a587362438e386bc81c14876a6684f0b53fc03d5e3d7d384a62ea54
semantic_inventory:
  - claim_key: CLAIM-5021B0234869E5FB
    claim_text: Authorization must be decided per account and be relationship-based
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-163#1
  recorded_at: '2026-11-01T22:44:56.700Z'
id: REQ-identity-authz-per-account
type: req
---
Authorization must be decided per account and be relationship-based.

## Context

The author decided it on 2026-11-01, when planning viewers for the second alpha (D-65). D-65 lets the owner enroll viewers from the second alpha and grant each the accounts they may view, so what a caller may do depends on its relationship to each account, not on a role. The read API answers per account (internal/views/app/access.go). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-163#1
