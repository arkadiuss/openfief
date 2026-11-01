---
title: An instrument the jurisdiction pack can't classify for D-127's exemptions must raise an attention item asking the owner to classify it once.
status: open
text_ref: design-decisions:D-132#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-132-2
  - lane:strict
semantic_text: An instrument the jurisdiction pack can't classify for D-127's exemptions must raise an attention item asking the owner to classify it once.
logic_claims:
  - CLAIM-710F90F9B08679BC
semantic_clauses:
  - An instrument the jurisdiction pack can't classify for D-127's exemptions must raise an attention item asking the owner to classify it once.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b2142865887e760c88ee412aa4067339a4afcc703dd03bcdc59f3fa8b404bf3e
semantic_inventory:
  - claim_key: CLAIM-710F90F9B08679BC
    claim_text: An instrument the jurisdiction pack can't classify for D-127's exemptions must raise an attention item asking the owner to classify it once
    role: normative
    status: modeled
    span:
      start: 0
      end: 139
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-132#2
  recorded_at: '2026-11-01T07:38:50.446Z'
id: REQ-valuation-unclassifiable-instrument-attention-item
type: req
---
An instrument the jurisdiction pack can't classify for D-127's exemptions must raise an attention item asking the owner to classify it once.

## Context

The author recorded D-132 on 2026-10-31 with OKI's yearly tax (D-127). The item goes to D-130's attention store (REQ-core-attention-item-store), and the owner answers it once per instrument. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, such as a fund whose policy it can't read, the instrument counts as neither, so the estimate errs high, and an attention item (D-130) asks the owner to classify it once; the owner's choice stays with the instrument.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-132#2
