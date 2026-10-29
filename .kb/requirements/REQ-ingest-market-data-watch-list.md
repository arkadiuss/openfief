---
title: An optional list in a market-data connector instance's config must add instruments to watch.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-93-3
  - lane:strict
text_ref: design-decisions:D-93#3
origin:
  kind: import
  ref: design-decisions:D-93#3
  recorded_at: '2026-10-29T08:48:02.750Z'
semantic_text: An optional list in a market-data connector instance's config must add instruments to watch.
logic_claims:
  - CLAIM-3D1062E339DB4D64
semantic_clauses:
  - An optional list in a market-data connector instance's config must add instruments to watch.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: f81383e896e7ea9a73ab4c2d572ede172e4c28f7de51cd200c259d78d47619d1
semantic_inventory:
  - claim_key: CLAIM-3D1062E339DB4D64
    claim_text: An optional list in a market-data connector instance's config must add instruments to watch
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
id: REQ-ingest-market-data-watch-list
type: req
---
An optional list in a market-data connector instance's config must add instruments to watch.

## Context

D-93 has each market-data sync carry the instruments the owner holds and the model portfolio's targets (REQ-ingest-market-data-sync-instruments). The owner may also want prices for instruments that are in neither, so D-48's per-instance config may hold an optional list of instruments to watch, which adds to them rather than replacing them. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> Each sync request to a market-data connector carries the instruments the owner holds and the model portfolio's targets, so a newly bought instrument is priced without configuration. An optional list in the instance's config adds instruments to watch.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-93#3
