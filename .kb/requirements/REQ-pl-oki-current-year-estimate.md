---
title: The current year's OKI tax must be an estimate over the days so far.
status: open
text_ref: design-decisions:D-146#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-146-2
  - lane:strict
semantic_text: The current year's OKI tax must be an estimate over the days so far.
logic_claims:
  - CLAIM-65D943A37C2E85EC
semantic_clauses:
  - The current year's OKI tax must be an estimate over the days so far.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 47e23adee9f38586a7742f424d4db7b693a177c48fe7bdf11c532ba7e3f09856
semantic_inventory:
  - claim_key: CLAIM-65D943A37C2E85EC
    claim_text: The current year's OKI tax must be an estimate over the days so far
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-146#2
  recorded_at: '2026-11-01T12:26:58.953Z'
id: REQ-pl-oki-current-year-estimate
type: req
---
The current year's OKI tax must be an estimate over the days so far.

## Context

The author recorded it on 2026-11-01, answering the implementation's OKI questions (#40 to #43, #46). OKI's tax is yearly (D-127), so before the year ends only the days so far are known. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An OKI account counts as held from its first transaction, and the current year's tax is an estimate over the days so far.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-146#2
