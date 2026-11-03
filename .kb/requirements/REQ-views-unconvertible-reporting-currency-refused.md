---
title: A reporting currency the jurisdiction pack can't convert to must stay refused until T-12 is resolved.
status: open
text_ref: design-decisions:D-180#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-180-4
  - lane:strict
semantic_text: A reporting currency the jurisdiction pack can't convert to must stay refused until T-12 is resolved.
logic_claims:
  - CLAIM-92921A478CABF9E3
semantic_clauses:
  - A reporting currency the jurisdiction pack can't convert to must stay refused until T-12 is resolved.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 17989aba62606f2f5a710064833375c3d2b602f7750ff7d3ad809310fec7f733
semantic_inventory:
  - claim_key: CLAIM-92921A478CABF9E3
    claim_text: A reporting currency the jurisdiction pack can't convert to must stay refused until T-12 is resolved
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-180#4
  recorded_at: '2026-11-02T23:54:35.201Z'
id: REQ-views-unconvertible-reporting-currency-refused
type: req
---
A reporting currency the jurisdiction pack can't convert to must stay refused until T-12 is resolved.

## Context

The author decided it on 2026-11-02, after reviewing where stored data assumed PLN (P0-6). T-12 records that NBP table A quotes every currency only against PLN, so another reporting currency needs cross rates, which are undecided (P-49). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#4
