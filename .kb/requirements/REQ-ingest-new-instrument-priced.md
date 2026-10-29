---
title: A newly bought instrument must be priced without configuration.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-93-2
  - lane:strict
text_ref: design-decisions:D-93#2
origin:
  kind: import
  ref: design-decisions:D-93#2
  recorded_at: '2026-10-29T08:47:27.741Z'
semantic_text: A newly bought instrument must be priced without configuration.
logic_claims:
  - CLAIM-A969843A686976D5
semantic_clauses:
  - A newly bought instrument must be priced without configuration.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 090587a065ea25b1471a6cb52dc4e21e697cfde6aa986fedba08526e4e4a31b1
semantic_inventory:
  - claim_key: CLAIM-A969843A686976D5
    claim_text: A newly bought instrument must be priced without configuration
    role: normative
    status: modeled
    span:
      start: 0
      end: 62
id: REQ-ingest-new-instrument-priced
type: req
---
A newly bought instrument must be priced without configuration.

## Context

D-93 gives the reason the core sends the instruments the owner holds on every market-data sync (REQ-ingest-market-data-sync-instruments): once a buy is imported, the next sync prices the new instrument, and the owner never has to add it to a connector's config first. The author recorded D-93 to D-95 after the design session that answered the implementation's questions.

## Source

> Each sync request to a market-data connector carries the instruments the owner holds and the model portfolio's targets, so a newly bought instrument is priced without configuration. An optional list in the instance's config adds instruments to watch.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-93#2
