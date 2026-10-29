---
title: A treasury-bonds connector sync must deliver, after its watermark, each final monthly CPI reading and, for each month whose reference-rate day is over, the NBP reference rate in force on that day, each dated on its month's first day, and return the last month of each as its watermark.
status: open
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
semantic_text: A treasury-bonds connector sync must deliver, after its watermark, each final monthly CPI reading and, for each month whose reference-rate day is over, the NBP reference rate in force on that day, each dated on its month's first day, and return the last month of each as its watermark.
logic_claims:
  - CLAIM-77C5223280767368
semantic_clauses:
  - A treasury-bonds connector sync must deliver, after its watermark, each final monthly CPI reading and, for each month whose reference-rate day is over, the NBP reference rate in force on that day, each dated on its month's first day, and return the last month of each as its watermark.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2d6e12a7a9fca2317ac589ba3834c043ce2b61dd24dd8eb2ecca90fb65ea0a17
semantic_inventory:
  - claim_key: CLAIM-77C5223280767368
    claim_text: A treasury-bonds connector sync must deliver, after its watermark, each final monthly CPI reading and, for each month whose reference-rate day is over, the NBP reference rate in force on that day, each dated on its month's first day, and return the last month of each as its watermark
    role: normative
    status: modeled
    span:
      start: 0
      end: 284
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:51.379Z'
id: REQ-treasurybonds-index-sync
type: req
---
A treasury-bonds connector sync must deliver, after its watermark, each final monthly CPI reading and, for each month whose reference-rate day is over, the NBP reference rate in force on that day, each dated on its month's first day, and return the last month of each as its watermark.

## Context

DESIGN.md's context diagram: Connector: treasury bonds, manual lots, CPI + NBP ref. rate. Valuation keys reference-rate readings by the month whose periods they set, so the connector picks the rate with the tenth-business-day rule and dates it on that month's first day; the contract's comment was changed to say so. The 13-year default start and four GUS years per sync are this task's choices. Modeled by Claude Code for the bond indices task; the author has not reviewed it yet.

## Source

> **[D-9]** M0 data paths: XTB through uploaded xStation export files; treasury bonds through manual lot entry (series, quantity, purchase date) with valuation computed; an import of the bond service's own export is added if one exists (OQ-22).

docs/DESIGN.md#D-9
