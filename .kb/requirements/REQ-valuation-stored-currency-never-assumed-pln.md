---
title: Neither a stored valuation's currency nor a stored exchange rate's base may be assumed to be PLN.
status: open
text_ref: design-decisions:D-180#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-180-3
  - lane:strict
semantic_text: Neither a stored valuation's currency nor a stored exchange rate's base may be assumed to be PLN.
logic_claims:
  - CLAIM-8FD22276C7E38385
semantic_clauses:
  - Neither a stored valuation's currency nor a stored exchange rate's base may be assumed to be PLN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d9ecd66173e9e8961e37ee0096c1ba2804280505a7112e6c483e9cb53b81074f
semantic_inventory:
  - claim_key: CLAIM-8FD22276C7E38385
    claim_text: Neither a stored valuation's currency nor a stored exchange rate's base may be assumed to be PLN
    role: descriptive
    status: modeled
    span:
      start: 0
      end: 96
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-180#3
  recorded_at: '2026-11-02T23:54:29.709Z'
id: REQ-valuation-stored-currency-never-assumed-pln
type: req
---
Neither a stored valuation's currency nor a stored exchange rate's base may be assumed to be PLN.

## Context

The author decided it on 2026-11-02, after reviewing where stored data assumed PLN (P0-6). D-62 says no code assumes PLN apart from the Polish pack's tax rules. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#3
