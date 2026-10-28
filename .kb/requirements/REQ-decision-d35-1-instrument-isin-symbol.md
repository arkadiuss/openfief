---
title: An instrument must have an ISIN with a valid check digit plus a venue symbol.
status: open
text_ref: design-decisions:D-35#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-35-1
  - lane:strict
semantic_text: An instrument must have an ISIN with a valid check digit plus a venue symbol.
logic_claims:
  - CLAIM-703BA0072C031906
semantic_clauses:
  - An instrument must have an ISIN with a valid check digit plus a venue symbol.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d9abfa075a0e1343636f7e3b6c82d98c5659496efc376df1fd6217f2661eee32
semantic_inventory:
  - claim_key: CLAIM-703BA0072C031906
    claim_text: An instrument must have an ISIN with a valid check digit plus a venue symbol
    role: normative
    status: modeled
    span:
      start: 0
      end: 76
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-35#1
  recorded_at: '2026-10-28T06:28:39.778Z'
id: REQ-decision-d35-1-instrument-isin-symbol
type: req
---
An instrument must have an ISIN with a valid check digit plus a venue symbol.

## Context

The author decided D-35 in docs/DESIGN.md section 3.3; this requirement imports its claim D-35#1 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An instrument needs an ISIN with a valid check digit plus a venue symbol.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-35#1
