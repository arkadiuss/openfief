---
title: Settings must show the audit log to the owner, filterable.
status: open
tags:
  - strict-modeling
  - lane:strict
  - provenance:design-decisions-d-211-5
text_ref: design-decisions:D-211#5
origin:
  kind: import
  ref: design-decisions:D-211#5
  recorded_at: '2026-11-12T19:18:18.474Z'
semantic_text: Settings must show the audit log to the owner, filterable.
logic_claims:
  - CLAIM-F2D838FA05873D93
semantic_clauses:
  - Settings must show the audit log to the owner, filterable.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 30fa4044defc81449034bfbbfd0c5174601da18c85196f28e942a315aab40ab3
semantic_inventory:
  - claim_key: CLAIM-F2D838FA05873D93
    claim_text: Settings must show the audit log to the owner, filterable
    role: normative
    status: modeled
    span:
      start: 0
      end: 57
id: REQ-identity-audit-owner-view
type: req
---
Settings must show the audit log to the owner, filterable.

## Context

The author decided it with D-211. This slice builds only the owner-only ListAuditEntries call, filtered by actor, action and time and paged, over the AuthService; the Settings page that shows it is a later slice, so this requirement stays open and its evidence floor is asserted.

## Source

> Settings shows it to the owner, filterable.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-211#5