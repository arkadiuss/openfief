---
title: Each sync request to a market-data connector must carry the instruments the owner holds and the model portfolio's targets.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-93-1
  - lane:strict
text_ref: design-decisions:D-93#1
origin:
  kind: import
  ref: design-decisions:D-93#1
  recorded_at: '2026-10-29T08:46:54.279Z'
semantic_text: Each sync request to a market-data connector must carry the instruments the owner holds and the model portfolio's targets.
logic_claims:
  - CLAIM-D3BA01BB08BEB655
semantic_clauses:
  - Each sync request to a market-data connector must carry the instruments the owner holds and the model portfolio's targets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: be0ae03c3f4f1e8052bf453ecf3a53b4a796fd8a6c8900b98532010b944d9fb1
semantic_inventory:
  - claim_key: CLAIM-D3BA01BB08BEB655
    claim_text: Each sync request to a market-data connector must carry the instruments the owner holds and the model portfolio's targets
    role: normative
    status: modeled
    span:
      start: 0
      end: 121
id: REQ-ingest-market-data-sync-instruments
type: req
---
Each sync request to a market-data connector must carry the instruments the owner holds and the model portfolio's targets.

## Context

D-49 has the core call connectors and own their schedules, and D-51 makes EODHD the first market-data connector. D-93 has the core tell a market-data connector what to price on every sync, the instruments the owner holds and the model portfolio's targets (D-52), so that a newly bought instrument is priced without configuration (REQ-ingest-new-instrument-priced); REQ-ingest-market-data-watch-list keeps an optional list of extra instruments in the instance's config. REQ-plugins-eodhd-config and REQ-plugins-eodhd-sync, which the author has not reviewed, still take every instrument from the instance's config. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> Each sync request to a market-data connector carries the instruments the owner holds and the model portfolio's targets, so a newly bought instrument is priced without configuration. An optional list in the instance's config adds instruments to watch.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-93#1
