---
title: An unclassified instrument's attention item must resolve itself once the owner classifies the instrument.
status: open
text_ref: design-decisions:D-148#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-148-2
  - lane:strict
semantic_text: An unclassified instrument's attention item must resolve itself once the owner classifies the instrument.
logic_claims:
  - CLAIM-0B335CC8D431D1B3
semantic_clauses:
  - An unclassified instrument's attention item must resolve itself once the owner classifies the instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e37451ecc3daa9f0c3b5eb2c1fa18f53904710466400c71f43d536d061c03b07
semantic_inventory:
  - claim_key: CLAIM-0B335CC8D431D1B3
    claim_text: An unclassified instrument's attention item must resolve itself once the owner classifies the instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 104
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-148#2
  recorded_at: '2026-11-01T12:27:44.610Z'
id: REQ-attention-unclassified-instrument-resolves-itself
type: req
---
An unclassified instrument's attention item must resolve itself once the owner classifies the instrument.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). The owner's classification stays with the instrument (REQ-valuation-owner-classification-stays), so the check finds nothing left to ask. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An unclassified instrument's attention item (D-132) is a check the attention store runs, so it resolves itself once the owner classifies the instrument.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-148#2
