---
title: A holding must link down to its instrument, and an allocation bar to the holdings behind it.
status: open
text_ref: design-decisions:D-195#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-195-3
  - lane:strict
semantic_text: A holding must link down to its instrument, and an allocation bar to the holdings behind it.
logic_claims:
  - CLAIM-9932253FBA2CE0B6
semantic_clauses:
  - A holding must link down to its instrument, and an allocation bar to the holdings behind it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ad4ba94e265765568ab6548fc7d571d611ca2fc96bd49940e13278bd6f189b98
semantic_inventory:
  - claim_key: CLAIM-9932253FBA2CE0B6
    claim_text: A holding must link down to its instrument, and an allocation bar to the holdings behind it
    role: normative
    status: modeled
    span:
      start: 0
      end: 91
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-195#3
  recorded_at: '2026-11-06T10:43:13.966Z'
id: REQ-decision-d195-3-holding-and-bar-link-down
type: req
---
A holding must link down to its instrument, and an allocation bar to the holdings behind it.

## Context

Author decision D-195 (UI v2 design session).

## Source

> **[D-195]** Every figure links down to what it is made of: the total to its accounts, an account to its holdings, a holding to its instrument, an allocation bar to the holdings behind it. An instrument has its own page: its price history and source, its holdings and lots across accounts, its transactions, its price checks, its ticker and ISIN, and its OKI class.

docs/DESIGN.md#D-195
