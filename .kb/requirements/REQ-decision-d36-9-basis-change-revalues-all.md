---
title: Changing the bond valuation basis must revalue everything.
status: open
text_ref: design-decisions:D-36#9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-9
  - lane:strict
semantic_text: Changing the bond valuation basis must revalue everything.
logic_claims:
  - CLAIM-0F3175C58C075AC5
semantic_clauses:
  - Changing the bond valuation basis must revalue everything.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5c338e1cbb7ce8a0458bfcade97c28fc5e71b3f18b49bf5c050eecdaa56f3376
semantic_inventory:
  - claim_key: CLAIM-0F3175C58C075AC5
    claim_text: Changing the bond valuation basis must revalue everything
    role: normative
    status: modeled
    span:
      start: 0
      end: 57
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#9
  recorded_at: '2026-10-28T06:30:20.913Z'
id: REQ-decision-d36-9-basis-change-revalues-all
type: req
---
Changing the bond valuation basis must revalue everything.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#9 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Configurable: worker concurrency, the end-of-day schedule and the bond basis (D-25); changing the basis revalues everything.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#9
