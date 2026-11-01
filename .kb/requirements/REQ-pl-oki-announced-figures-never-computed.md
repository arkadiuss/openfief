---
title: OpenFief must never compute OKI's rate from 2028 or its exempt amounts from 2030.
status: open
text_ref: design-decisions:D-145#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-145-3
  - lane:strict
semantic_text: OpenFief must never compute OKI's rate from 2028 or its exempt amounts from 2030.
logic_claims:
  - CLAIM-AAF9C3965904CA3E
semantic_clauses:
  - OpenFief must never compute OKI's rate from 2028 or its exempt amounts from 2030.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6410dbe679bd5c27e9ed8ff0c4d0cefaebc459812fe1f5f11e4149340ae48759
semantic_inventory:
  - claim_key: CLAIM-AAF9C3965904CA3E
    claim_text: OpenFief must never compute OKI's rate from 2028 or its exempt amounts from 2030
    role: exception
    status: modeled
    span:
      start: 0
      end: 80
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-145#3
  recorded_at: '2026-11-01T12:26:32.004Z'
id: REQ-pl-oki-announced-figures-never-computed
type: req
---
OpenFief must never compute OKI's rate from 2028 or its exempt amounts from 2030.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). The figures arrive with each release in the act's data (REQ-pl-oki-rate-from-announcements, REQ-pl-oki-exempt-amounts-from-announcements). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> OKI's rate from 2028 and its exempt amounts from 2030 come from the finance minister's announcements, added to the act's data with each release; OpenFief never computes them, and the amounts never fall.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-145#3
