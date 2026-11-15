---
title: A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-223#1
tags:
  - views
  - period
  - performance
  - d-223
  - d-61
  - d-138
origin:
  kind: import
  ref: design-decisions:D-223#1
  recorded_at: '2026-11-15T18:11:37.441Z'
semantic_text: A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9aeca525b2983dd487e46eb404f5e49cc90dfc967c0828b51fa3ae98932148d6
logic_claims:
  - CLAIM-F8BA46C58CEF26A3
semantic_inventory:
  - claim_key: CLAIM-F8BA46C58CEF26A3
    claim_text: A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 135
id: REQ-views-period-return-from-opening-close
type: req
---
A period's return must run from the close of the day before its first day to the close of its last day, so its first day counts in full.

## Context

The author decided it as D-223 on 2026-11-15, after the Ask box's period cards (REQ-ui-ask-every-period-one-card) put every period on one convention: 2025's return runs from the close of 31 December 2024. D-61 chains the time-weighted return across cash flows and D-138 puts the performance index at 100 on a chart's start, which had made a chart opened on a period's first day leave that day out. The core's GetPeriodChange computes it, from the opening day it names. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> **[D-223]** A period's return runs from the close of the day before its first day to the close of its last day, so its first day counts in full: 2025's runs from the close of 31 December 2024.

docs/DESIGN.md#D-223, D-61, D-138