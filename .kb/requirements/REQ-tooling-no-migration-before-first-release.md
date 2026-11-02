---
title: Data stored before the first release must not be migrated.
status: open
text_ref: design-decisions:D-161#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-161-1
  - lane:strict
semantic_text: Data stored before the first release must not be migrated.
logic_claims:
  - CLAIM-ED852AE4D402D5A3
semantic_clauses:
  - Data stored before the first release must not be migrated.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ab1143b3c59d4aba6650ea71d3881a1369824a14b433689e8d158c3ade39d098
semantic_inventory:
  - claim_key: CLAIM-ED852AE4D402D5A3
    claim_text: Data stored before the first release must not be migrated
    role: normative
    status: modeled
    span:
      start: 0
      end: 57
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-161#1
  recorded_at: '2026-11-02T00:54:56.055Z'
id: REQ-tooling-no-migration-before-first-release
type: req
---
Data stored before the first release must not be migrated.

## Context

The author recorded it on 2026-11-01. Reason not stated. The decision accepts that a change before the first release may need data entered again. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Nothing stored before the first release is migrated: a change before it may need data entered again, and the demo is reseeded. From the first release on, every change to stored data ships with a migration of what is already stored.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-161#1
