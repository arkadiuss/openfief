---
title: An unclassified instrument's attention item must be a check the attention store runs.
status: open
text_ref: design-decisions:D-148#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-148-1
  - lane:strict
semantic_text: An unclassified instrument's attention item must be a check the attention store runs.
logic_claims:
  - CLAIM-DC95E053F5D2AFC3
semantic_clauses:
  - An unclassified instrument's attention item must be a check the attention store runs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9dcb9d19e6206cfd46f9d9a70f582741a20a60c464538ff455fbf59bd9a39da4
semantic_inventory:
  - claim_key: CLAIM-DC95E053F5D2AFC3
    claim_text: An unclassified instrument's attention item must be a check the attention store runs
    role: normative
    status: modeled
    span:
      start: 0
      end: 84
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-148#1
  recorded_at: '2026-11-01T12:27:33.224Z'
id: REQ-attention-unclassified-instrument-is-check
type: req
---
An unclassified instrument's attention item must be a check the attention store runs.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). D-132 has an instrument the pack can't classify raise an attention item (REQ-valuation-unclassifiable-instrument-attention-item), in D-130's store. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An unclassified instrument's attention item (D-132) is a check the attention store runs, so it resolves itself once the owner classifies the instrument.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-148#1
