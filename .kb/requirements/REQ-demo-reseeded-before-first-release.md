---
title: Before the first release, the demo must be reseeded rather than its stored data migrated.
status: open
text_ref: design-decisions:D-161#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-161-2
  - lane:strict
semantic_text: Before the first release, the demo must be reseeded rather than its stored data migrated.
logic_claims:
  - CLAIM-E0AB1C52734A5F74
semantic_clauses:
  - Before the first release, the demo must be reseeded rather than its stored data migrated.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 26a3c58435021f95af26486e51fd77bf1a25345c088b3d25203842e1228f6c68
semantic_inventory:
  - claim_key: CLAIM-E0AB1C52734A5F74
    claim_text: Before the first release, the demo must be reseeded rather than its stored data migrated
    role: normative
    status: modeled
    span:
      start: 0
      end: 88
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-161#2
  recorded_at: '2026-11-02T00:55:02.308Z'
id: REQ-demo-reseeded-before-first-release
type: req
---
Before the first release, the demo must be reseeded rather than its stored data migrated.

## Why this changed

On 2026-11-01 the author recorded D-161. REQ-demo-reseed-on-new-loader, a default requirement the implementation agent recorded while building D-161 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-161 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The demo's dataset is synthetic and loaded by the demo loader (docs/dev/demo.md). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Nothing stored before the first release is migrated: a change before it may need data entered again, and the demo is reseeded. From the first release on, every change to stored data ships with a migration of what is already stored.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-161#2
