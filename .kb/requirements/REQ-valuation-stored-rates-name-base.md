---
title: Stored exchange rates must name their base currency, taken from the jurisdiction pack.
status: open
text_ref: design-decisions:D-180#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-180-2
  - lane:strict
semantic_text: Stored exchange rates must name their base currency, taken from the jurisdiction pack.
logic_claims:
  - CLAIM-38A0161B64A0FE2B
semantic_clauses:
  - Stored exchange rates must name their base currency, taken from the jurisdiction pack.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a25a76ec3cf3bc14ad7f24a836858d0c09ff4455b1758fec24de9256d218eb4d
semantic_inventory:
  - claim_key: CLAIM-38A0161B64A0FE2B
    claim_text: Stored exchange rates must name their base currency, taken from the jurisdiction pack
    role: normative
    status: modeled
    span:
      start: 0
      end: 85
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-180#2
  recorded_at: '2026-11-02T23:54:21.886Z'
id: REQ-valuation-stored-rates-name-base
type: req
---
Stored exchange rates must name their base currency, taken from the jurisdiction pack.

## Context

The author decided it on 2026-11-02, after reviewing where stored data assumed PLN (P0-6). D-20 converts amounts with NBP table A, which quotes against PLN, behind a swappable rate provider. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#2
