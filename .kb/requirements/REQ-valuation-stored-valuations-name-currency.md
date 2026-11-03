---
title: Stored valuations must name their currency, taken from the jurisdiction pack.
status: open
text_ref: design-decisions:D-180#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-180-1
  - lane:strict
semantic_text: Stored valuations must name their currency, taken from the jurisdiction pack.
logic_claims:
  - CLAIM-C420719407131C16
semantic_clauses:
  - Stored valuations must name their currency, taken from the jurisdiction pack.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3a9e889d4c3e5066599a70d9fe8b517b80097550b4077fdbf511f77595fb0507
semantic_inventory:
  - claim_key: CLAIM-C420719407131C16
    claim_text: Stored valuations must name their currency, taken from the jurisdiction pack
    role: normative
    status: modeled
    span:
      start: 0
      end: 76
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-180#1
  recorded_at: '2026-11-02T23:54:14.395Z'
id: REQ-valuation-stored-valuations-name-currency
type: req
---
Stored valuations must name their currency, taken from the jurisdiction pack.

## Context

The author decided it on 2026-11-02, after reviewing where stored data assumed PLN (P0-6). D-62 makes the reporting currency a setting whose default comes from the pack and has no code assume PLN apart from the Polish pack's tax rules; D-36 stores valuations. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1
