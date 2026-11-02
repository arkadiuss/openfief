---
title: From the first release on, every change to stored data must ship with a migration of what is already stored.
status: open
text_ref: design-decisions:D-161#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-161-3
  - lane:strict
semantic_text: From the first release on, every change to stored data must ship with a migration of what is already stored.
logic_claims:
  - CLAIM-21FA81B22FD78C4D
semantic_clauses:
  - From the first release on, every change to stored data must ship with a migration of what is already stored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ac0d6582c17f7568272079447d4f98767d6528ab01fbd27d1695abd0eb04b277
semantic_inventory:
  - claim_key: CLAIM-21FA81B22FD78C4D
    claim_text: From the first release on, every change to stored data must ship with a migration of what is already stored
    role: normative
    status: modeled
    span:
      start: 0
      end: 107
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-161#3
  recorded_at: '2026-11-02T00:55:07.877Z'
id: REQ-tooling-migration-from-first-release
type: req
---
From the first release on, every change to stored data must ship with a migration of what is already stored.

## Context

The author recorded it on 2026-11-01. Reason not stated. Each context migrates its own PostgreSQL schema (D-14). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Nothing stored before the first release is migrated: a change before it may need data entered again, and the demo is reseeded. From the first release on, every change to stored data ships with a migration of what is already stored.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-161#3
