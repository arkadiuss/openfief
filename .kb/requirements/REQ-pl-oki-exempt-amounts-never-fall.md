---
title: OKI's exempt amounts must never fall.
status: open
text_ref: design-decisions:D-145#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-145-4
  - lane:strict
semantic_text: OKI's exempt amounts must never fall.
logic_claims:
  - CLAIM-A3DB75BAC98B3429
semantic_clauses:
  - OKI's exempt amounts must never fall.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1c86c99d6ae49be00dd2e389f3f98e3379fd86886eb7290d9b7f01e73b87eeb8
semantic_inventory:
  - claim_key: CLAIM-A3DB75BAC98B3429
    claim_text: OKI's exempt amounts must never fall
    role: exception
    status: modeled
    span:
      start: 0
      end: 36
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-145#4
  recorded_at: '2026-11-01T12:26:40.839Z'
id: REQ-pl-oki-exempt-amounts-never-fall
type: req
---
OKI's exempt amounts must never fall.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). D-127's exemptions are 25,000 PLN for savings assets and 100,000 PLN together until they rise with CPI from 2030 (REQ-pl-oki-tax-exemptions). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> OKI's rate from 2028 and its exempt amounts from 2030 come from the finance minister's announcements, added to the act's data with each release; OpenFief never computes them, and the amounts never fall.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-145#4
