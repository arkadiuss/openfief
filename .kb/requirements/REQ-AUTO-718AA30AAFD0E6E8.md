---
title: The built-in UI must show a value-over-time chart that starts at go-live.
status: open
text_ref: design-decisions:D-30#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-30-2
  - lane:strict
semantic_text: The built-in UI must show a value-over-time chart that starts at go-live.
logic_claims:
  - CLAIM-D3F17F4743FDBABF
semantic_clauses:
  - The built-in UI must show a value-over-time chart that starts at go-live.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 287991c51171f7178d07b75d9bd9c64c95767cefb244c3bc985aad7eac500d36
semantic_inventory:
  - claim_key: CLAIM-D3F17F4743FDBABF
    claim_text: The built-in UI must show a value-over-time chart that starts at go-live
    role: normative
    status: modeled
    span:
      start: 0
      end: 72
origin:
  kind: import
  ref: design-decisions:D-30#2
  recorded_at: '2026-10-27T20:12:51.024Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-718AA30AAFD0E6E8
type: req
---
The built-in UI must show a value-over-time chart that starts at go-live.

## Context

D-30 names the charts the built-in UI draws in M0. It first named two, allocation and value over time; with D-106 the author separated value and performance, and D-30 now names three. The value-over-time chart shows money, cash included, with D-77's switch (D-106).

## Source

> The built-in UI draws three charts in M0: allocation (by asset class, currency, broker or sleeve), value over time, and performance (D-106). The time charts start at go-live (D-23).

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-30#2
