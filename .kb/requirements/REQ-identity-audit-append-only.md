---
title: The audit log must be append-only.
status: open
tags:
  - strict-modeling
  - lane:strict
  - provenance:design-decisions-d-211-1
text_ref: design-decisions:D-211#1
origin:
  kind: import
  ref: design-decisions:D-211#1
  recorded_at: '2026-11-12T19:17:54.248Z'
semantic_text: The audit log must be append-only.
logic_claims:
  - CLAIM-A9B6200DD8915BC5
semantic_clauses:
  - The audit log must be append-only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a9a11b9005826efbb19ed9447543a7b43387c53fccf9d2f8a466c32f84902095
semantic_inventory:
  - claim_key: CLAIM-A9B6200DD8915BC5
    claim_text: The audit log must be append-only
    role: normative
    status: modeled
    span:
      start: 0
      end: 33
id: REQ-identity-audit-append-only
type: req
---
The audit log must be append-only.

## Context

The author decided it with D-211, and in the design session before this slice chose that the database enforces it: the core's role gets no UPDATE, DELETE or TRUNCATE on the table and a trigger refuses them for any role, with no hash chain yet (P-18 is undecided) and retention forever. No end-to-end test proves it, so its evidence floor is asserted.

## Source

> An append-only audit log records every change of state and every access event.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-211#1