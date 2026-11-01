---
title: OKI's rate from 2028 must come from the finance minister's announcements, added to the act's data with each release.
status: open
text_ref: design-decisions:D-145#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-145-1
  - lane:strict
semantic_text: OKI's rate from 2028 must come from the finance minister's announcements, added to the act's data with each release.
logic_claims:
  - CLAIM-7D0C2E2FF637E74E
semantic_clauses:
  - OKI's rate from 2028 must come from the finance minister's announcements, added to the act's data with each release.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 25e6eee58c0917d3778f90196b13f9f040f01121279b84bf68ad9aedb25131c9
semantic_inventory:
  - claim_key: CLAIM-7D0C2E2FF637E74E
    claim_text: OKI's rate from 2028 must come from the finance minister's announcements, added to the act's data with each release
    role: normative
    status: modeled
    span:
      start: 0
      end: 115
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-145#1
  recorded_at: '2026-11-01T12:26:16.097Z'
id: REQ-pl-oki-rate-from-announcements
type: req
---
OKI's rate from 2028 must come from the finance minister's announcements, added to the act's data with each release.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). D-127 says how the act sets the rate after 2027 (REQ-pl-oki-tax-rate); the minister announces it. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> OKI's rate from 2028 and its exempt amounts from 2030 come from the finance minister's announcements, added to the act's data with each release; OpenFief never computes them, and the amounts never fall.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-145#1
