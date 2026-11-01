---
title: The demo loader must reseed a marked demo database that another build of the loader filled, so the demo's data always matches the code deployed, without migrating what an older build stored.
status: open
text_ref: docs/DESIGN.md#D-161
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-161
  - lane:strict
semantic_text: The demo loader must reseed a marked demo database that another build of the loader filled, so the demo's data always matches the code deployed, without migrating what an older build stored.
logic_claims:
  - CLAIM-D1B8401805C0237B
semantic_clauses:
  - The demo loader must reseed a marked demo database that another build of the loader filled, so the demo's data always matches the code deployed, without migrating what an older build stored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b756304fefb6ff84df8342d8ec2cf1138ee40e65c013cb84671a71c49cbd19af
semantic_inventory:
  - claim_key: CLAIM-D1B8401805C0237B
    claim_text: The demo loader must reseed a marked demo database that another build of the loader filled, so the demo's data always matches the code deployed, without migrating what an older build stored
    role: normative
    status: modeled
    span:
      start: 0
      end: 189
origin:
  kind: agent
  recorded_at: '2026-11-01T16:51:03.533Z'
id: REQ-demo-reseed-on-new-loader
type: req
---
The demo loader must reseed a marked demo database that another build of the loader filled, so the demo's data always matches the code deployed, without migrating what an older build stored.

## Context

D-161 migrates nothing stored before the first release and reseeds the demo instead. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Source

> Nothing stored before the first release is migrated: a change before it may need data entered again, and the demo is reseeded. From the first release on, every change to stored data ships with a migration of what is already stored.

docs/DESIGN.md#D-161
