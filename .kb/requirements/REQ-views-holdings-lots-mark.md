---
title: A holding must be marked as no longer updated when any current transaction posting to it on or before the date came from a removed connector instance, and a lot when the transaction that opened it or a disposal that consumed part of it did, naming the instances.
status: open
text_ref: docs/DESIGN.md#D-91
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-91
  - lane:strict
semantic_text: A holding must be marked as no longer updated when any current transaction posting to it on or before the date came from a removed connector instance, and a lot when the transaction that opened it or a disposal that consumed part of it did, naming the instances.
logic_claims:
  - CLAIM-810F49BC47A047D8
semantic_clauses:
  - A holding must be marked as no longer updated when any current transaction posting to it on or before the date came from a removed connector instance, and a lot when the transaction that opened it or a disposal that consumed part of it did, naming the instances.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0f4634b5ea502e6f28b7534a0e3554a2c4d2e4278590c822ef54e533351686ea
semantic_inventory:
  - claim_key: CLAIM-810F49BC47A047D8
    claim_text: A holding must be marked as no longer updated when any current transaction posting to it on or before the date came from a removed connector instance, and a lot when the transaction that opened it or a disposal that consumed part of it did, naming the instances
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-10-30T00:05:20.894Z'
id: REQ-views-holdings-lots-mark
type: req
---
A holding must be marked as no longer updated when any current transaction posting to it on or before the date came from a removed connector instance, and a lot when the transaction that opened it or a disposal that consumed part of it did, naming the instances.

## Context

D-91 marks holdings and lots behind a removed instance, worked out when read. Which transactions are behind a lot, its opening one and the disposals that cut it, is this package's reading. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-91]** Holdings and lots are marked as no longer updated when any transaction behind them comes from a removed instance. Like the transactions' own mark, it is worked out when the data is read, so the log stays append-only.

docs/DESIGN.md#D-91
