---
title: An OKI account must count as held from its first transaction.
status: open
text_ref: design-decisions:D-146#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-146-1
  - lane:strict
semantic_text: An OKI account must count as held from its first transaction.
logic_claims:
  - CLAIM-4CCDB623EEAE580F
semantic_clauses:
  - An OKI account must count as held from its first transaction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7f0e99308654c2c055a7ebc6ad2ab0b4a3d03e9ee5c6f40585551483f549614f
semantic_inventory:
  - claim_key: CLAIM-4CCDB623EEAE580F
    claim_text: An OKI account must count as held from its first transaction
    role: normative
    status: modeled
    span:
      start: 0
      end: 60
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-146#1
  recorded_at: '2026-11-01T12:26:49.890Z'
id: REQ-pl-oki-account-held-from-first-transaction
type: req
---
An OKI account must count as held from its first transaction.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). D-127 divides each account's average value by the days of the year it was held (REQ-pl-oki-tax-base-average-value). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An OKI account counts as held from its first transaction, and the current year's tax is an estimate over the days so far.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-146#1
