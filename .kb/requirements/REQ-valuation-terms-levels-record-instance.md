---
title: Every stored bond series term and market index level must record the connector instance that delivered it.
status: open
text_ref: design-decisions:D-179#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-179-4
  - lane:strict
semantic_text: Every stored bond series term and market index level must record the connector instance that delivered it.
logic_claims:
  - CLAIM-F78CF981C4A7F433
semantic_clauses:
  - Every stored bond series term and market index level must record the connector instance that delivered it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3ebb627b9dc662c0b7a255ec472f479ba7c4cafd8da28f23df70ceddaf336225
semantic_inventory:
  - claim_key: CLAIM-F78CF981C4A7F433
    claim_text: Every stored bond series term and market index level must record the connector instance that delivered it
    role: normative
    status: modeled
    span:
      start: 0
      end: 105
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-179#4
  recorded_at: '2026-11-02T22:24:29.439Z'
id: REQ-valuation-terms-levels-record-instance
type: req
---
Every stored bond series term and market index level must record the connector instance that delivered it.

## Context

The author decided it on 2026-11-02, after reviewing what a connector may deliver into the stores that value holdings (P0-5). D-124 makes a source one connector instance; prices, FX rates and index readings already keep theirs (D-94). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them; anything else it delivers of that kind is dropped with an attention item (D-130), as D-141 does for cross-check prices. Bond terms need a Polish (`PL`) ISIN, and every stored term and index level records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-179#4
