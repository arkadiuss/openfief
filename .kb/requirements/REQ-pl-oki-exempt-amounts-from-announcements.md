---
title: OKI's exempt amounts from 2030 must come from the finance minister's announcements, added to the act's data with each release.
status: open
text_ref: design-decisions:D-145#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-145-2
  - lane:strict
semantic_text: OKI's exempt amounts from 2030 must come from the finance minister's announcements, added to the act's data with each release.
logic_claims:
  - CLAIM-063BCFF0B9CED775
semantic_clauses:
  - OKI's exempt amounts from 2030 must come from the finance minister's announcements, added to the act's data with each release.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d6203a890748ec2ae88f04df884741de71bdef97e652b634f0fbe2c297276aea
semantic_inventory:
  - claim_key: CLAIM-063BCFF0B9CED775
    claim_text: OKI's exempt amounts from 2030 must come from the finance minister's announcements, added to the act's data with each release
    role: exception
    status: modeled
    span:
      start: 0
      end: 125
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-145#2
  recorded_at: '2026-11-01T12:26:24.024Z'
id: REQ-pl-oki-exempt-amounts-from-announcements
type: req
---
OKI's exempt amounts from 2030 must come from the finance minister's announcements, added to the act's data with each release.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). D-127 has the exemptions rise with CPI from 2030 (REQ-pl-oki-exemptions-cpi-indexed); the minister announces the amounts. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> OKI's rate from 2028 and its exempt amounts from 2030 come from the finance minister's announcements, added to the act's data with each release; OpenFief never computes them, and the amounts never fall.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-145#2
