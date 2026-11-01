---
title: The built-in UI must show a performance chart that starts on the charts' start date.
status: open
text_ref: design-decisions:D-30#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-30-3
  - lane:strict
semantic_text: The built-in UI must show a performance chart that starts on the charts' start date.
logic_claims:
  - CLAIM-DC4BFE08DBF23DE5
semantic_clauses:
  - The built-in UI must show a performance chart that starts on the charts' start date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f672d581fd1017e5fafe53abeb79f6b6b29280f10ef36a719b717656486c172e
semantic_inventory:
  - claim_key: CLAIM-DC4BFE08DBF23DE5
    claim_text: The built-in UI must show a performance chart that starts on the charts' start date
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-30#3
  recorded_at: '2026-11-01T07:38:22.015Z'
id: REQ-ui-performance-chart-from-charts-start
type: req
---
The built-in UI must show a performance chart that starts on the charts' start date.

## Why this changed

On 2026-11-01 the author recorded D-138 and changed D-30, D-53 and D-106 to start the time charts on the charts' start (D-138) instead of go-live. With real purchase dates (D-98) keeping the ledger a record of what happened, where a chart starts is the owner's choice, with the performance index at 100 on it. REQ-ui-performance-chart started the performance chart at go-live, so this requirement replaces it.

## Context

D-30 lists the built-in UI's three M0 charts and D-106 says what the performance chart shows; D-138 says how the charts' start is chosen. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The built-in UI draws three charts in M0: allocation (by asset class, currency, broker or sleeve), value over time, and performance (D-106). The time charts start on the charts' start date (D-138).

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-30#3
