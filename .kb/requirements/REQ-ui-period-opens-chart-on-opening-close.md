---
title: Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-223#2
tags:
  - ui
  - performance
  - period
  - d-223
  - d-138
  - d-61
origin:
  kind: import
  ref: design-decisions:D-223#2
  recorded_at: '2026-11-15T18:11:38.733Z'
semantic_text: Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dc89ce03b35413b70d739d4560d76c0c453af4d55d8e3b4d1a87ed9e468ad7b7
logic_claims:
  - CLAIM-F5005365120F0539
semantic_inventory:
  - claim_key: CLAIM-F5005365120F0539
    claim_text: Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 242
id: REQ-ui-period-opens-chart-on-opening-close
type: req
---
Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start must be set to the period's opening close, so that the chart, which keeps 100 on its start, and the card show the same return.

## Context

The author decided it as D-223 on 2026-11-15. D-138 and REQ-ui-performance-index-100-on-start keep the performance index at 100 on the charts' start, so a chart that starts on a period's first day shows a return that leaves that day's move out, and differs from the period card's, which runs from the close of the day before. Setting the start to the opening close, rather than changing the chart's convention, makes the two agree. A start the owner types by hand stays as typed. The overview has no preset row: the one link that opens a period on the chart besides the Ask box's card is the value card's time-weighted return of this year, which opens at the year's opening close. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> **[D-223]** Wherever a period opens on the performance chart, from one of its presets or from an Ask box card, the charts' start (D-138) is set to the period's opening close, so the chart, which keeps 100 on its start, and the card show the same return.

docs/DESIGN.md#D-223, D-138, D-61