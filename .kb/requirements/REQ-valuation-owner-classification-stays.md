---
title: The owner's classification of an instrument for D-127's exemptions must stay with the instrument.
status: open
text_ref: design-decisions:D-132#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-132-3
  - lane:strict
semantic_text: The owner's classification of an instrument for D-127's exemptions must stay with the instrument.
logic_claims:
  - CLAIM-EEAC916402904506
semantic_clauses:
  - The owner's classification of an instrument for D-127's exemptions must stay with the instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d73684d0559e8f5f8deee92b2270e95fd9af470b0211c0deefbd577cedbe4bca
semantic_inventory:
  - claim_key: CLAIM-EEAC916402904506
    claim_text: The owner's classification of an instrument for D-127's exemptions must stay with the instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-132#3
  recorded_at: '2026-11-01T07:38:56.525Z'
id: REQ-valuation-owner-classification-stays
type: req
---
The owner's classification of an instrument for D-127's exemptions must stay with the instrument.

## Context

The author recorded D-132 on 2026-10-31 with OKI's yearly tax (D-127), so the owner classifies an instrument the pack can't once, and the choice holds for every account and later year that holds it. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, such as a fund whose policy it can't read, the instrument counts as neither, so the estimate errs high, and an attention item (D-130) asks the owner to classify it once; the owner's choice stays with the instrument.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-132#3
