---
title: When the jurisdiction pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, the instrument must count as neither.
status: open
text_ref: design-decisions:D-132#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-132-1
  - lane:strict
semantic_text: When the jurisdiction pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, the instrument must count as neither.
logic_claims:
  - CLAIM-359A4C85E3FF9B20
semantic_clauses:
  - When the jurisdiction pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, the instrument must count as neither.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 735fcb2574ae365085aa04f1fccc801b0b7ac4578c7f1f3e1a00ef28fbf27b6f
semantic_inventory:
  - claim_key: CLAIM-359A4C85E3FF9B20
    claim_text: When the jurisdiction pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, the instrument must count as neither
    role: condition
    status: modeled
    span:
      start: 0
      end: 171
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-132#1
  recorded_at: '2026-11-01T07:38:44.208Z'
id: REQ-valuation-unclassifiable-instrument-counts-as-neither
type: req
---
When the jurisdiction pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, the instrument must count as neither.

## Context

The author recorded D-132 on 2026-10-31 with OKI's yearly tax (D-127). The decision gives the reason: counting such an instrument as neither makes the estimate err high. A fund whose policy the pack can't read is the example. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> When the pack can't tell whether an instrument is a savings asset, an investment asset or neither for D-127's exemptions, such as a fund whose policy it can't read, the instrument counts as neither, so the estimate errs high, and an attention item (D-130) asks the owner to classify it once; the owner's choice stays with the instrument.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-132#1
