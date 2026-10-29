---
title: The Ghostfolio plugin must create its Ghostfolio accounts with UUIDv5 identifiers derived from the plugin instance and the OpenFief account, so it knows which accounts it owns without storing anything.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-5
  - lane:strict
text_ref: design-decisions:D-64#5
origin:
  kind: import
  ref: design-decisions:D-64#5
  recorded_at: '2026-10-29T10:22:48.867Z'
semantic_text: The Ghostfolio plugin must create its Ghostfolio accounts with UUIDv5 identifiers derived from the plugin instance and the OpenFief account, so it knows which accounts it owns without storing anything.
logic_claims:
  - CLAIM-3BB81D4F9090DF49
semantic_clauses:
  - The Ghostfolio plugin must create its Ghostfolio accounts with UUIDv5 identifiers derived from the plugin instance and the OpenFief account, so it knows which accounts it owns without storing anything.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 71a0cba4dd3ec4b4f0ef8aa7ee95ee47db367adb71d59da823dc33369dc8bd6c
semantic_inventory:
  - claim_key: CLAIM-3BB81D4F9090DF49
    claim_text: The Ghostfolio plugin must create its Ghostfolio accounts with UUIDv5 identifiers derived from the plugin instance and the OpenFief account, so it knows which accounts it owns without storing anything
    role: normative
    status: modeled
    span:
      start: 0
      end: 200
id: REQ-views-ghostfolio-derived-account-uuids
type: req
---
The Ghostfolio plugin must create its Ghostfolio accounts with UUIDv5 identifiers derived from the plugin instance and the OpenFief account, so it knows which accounts it owns without storing anything.

## Context

D-64 first had the plugin generate and store the UUIDs of its Ghostfolio accounts. The author amended D-64 to follow the plugin as built: the UUIDs are derived instead, because a plugin has no volume and no Kubernetes API access to store them in. The author amended D-64 and recorded D-108 after the design session that answered the implementation's questions.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - It creates its Ghostfolio accounts with UUIDs derived from the plugin instance and the OpenFief account (UUIDv5), so it knows which accounts it owns without storing anything: a plugin has no volume and no Kubernetes API access.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#5
