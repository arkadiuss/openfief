---
title: The Ghostfolio plugin must take each listed instrument's Ghostfolio data source and symbol from the owner's mapping of its ISIN.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-108-4
  - lane:strict
text_ref: design-decisions:D-108#4
origin:
  kind: import
  ref: design-decisions:D-108#4
  recorded_at: '2026-10-29T10:27:16.743Z'
semantic_text: The Ghostfolio plugin must take each listed instrument's Ghostfolio data source and symbol from the owner's mapping of its ISIN.
logic_claims:
  - CLAIM-47E71F612B6D25F2
semantic_clauses:
  - The Ghostfolio plugin must take each listed instrument's Ghostfolio data source and symbol from the owner's mapping of its ISIN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0a9ddec8f9076b78d7723677e8c3d9952348997f8cb2955346f4eb3f7680336e
semantic_inventory:
  - claim_key: CLAIM-47E71F612B6D25F2
    claim_text: The Ghostfolio plugin must take each listed instrument's Ghostfolio data source and symbol from the owner's mapping of its ISIN
    role: normative
    status: modeled
    span:
      start: 0
      end: 127
id: REQ-views-ghostfolio-isin-mapping
type: req
---
The Ghostfolio plugin must take each listed instrument's Ghostfolio data source and symbol from the owner's mapping of its ISIN.

## Context

D-35 identifies an instrument by its ISIN and a venue symbol, while Ghostfolio identifies a listed asset by its own data source and symbol. D-108 records, as the author's decision, that the plugin as built needs the owner's mapping between the two. The author amended D-64 and recorded D-108 after the design session that answered the implementation's questions.

## Source

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-108#4
