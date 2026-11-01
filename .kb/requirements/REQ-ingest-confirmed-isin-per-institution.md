---
title: A ticker's confirmed ISIN must hold for every account at the same institution.
status: open
text_ref: design-decisions:D-153#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-153-1
  - lane:strict
semantic_text: A ticker's confirmed ISIN must hold for every account at the same institution.
logic_claims:
  - CLAIM-F60B424030C013F0
semantic_clauses:
  - A ticker's confirmed ISIN must hold for every account at the same institution.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fc77c97e73eae8417debe2c888941522ab9c46c2c104b4cb6e6b65aac666c1bb
semantic_inventory:
  - claim_key: CLAIM-F60B424030C013F0
    claim_text: A ticker's confirmed ISIN must hold for every account at the same institution
    role: normative
    status: modeled
    span:
      start: 0
      end: 77
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-153#1
  recorded_at: '2026-11-01T12:29:43.224Z'
id: REQ-ingest-confirmed-isin-per-institution
type: req
---
A ticker's confirmed ISIN must hold for every account at the same institution.

## Context

The author recorded it on 2026-11-01. D-116 has the owner confirm a new XTB ticker's ISIN once (REQ-ingest-xtb-isin-owner-confirms); this says that once covers every account at that institution. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A ticker's confirmed ISIN (D-116) holds for every account at the same institution. The owner can correct a wrong one, which supersedes the transactions it booked (D-18).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-153#1
