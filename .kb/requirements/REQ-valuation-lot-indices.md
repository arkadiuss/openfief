---
title: A bond lot's interest periods after the first that have started must take the CPI reading of the month two months before the period starts for inflation-indexed types, and the reference-rate reading recorded for the month the period starts in for ROR and DOR; fixed-rate types need none.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: A bond lot's interest periods after the first that have started must take the CPI reading of the month two months before the period starts for inflation-indexed types, and the reference-rate reading recorded for the month the period starts in for ROR and DOR; fixed-rate types need none.
logic_claims:
  - CLAIM-E24B54D799CAAE25
semantic_clauses:
  - A bond lot's interest periods after the first that have started must take the CPI reading of the month two months before the period starts for inflation-indexed types, and the reference-rate reading recorded for the month the period starts in for ROR and DOR; fixed-rate types need none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8f6f2ef661d1eb86ad6461e7a98414ba490436cdb426b1e1b27ef053d38c07e7
semantic_inventory:
  - claim_key: CLAIM-E24B54D799CAAE25
    claim_text: A bond lot's interest periods after the first that have started must take the CPI reading of the month two months before the period starts for inflation-indexed types, and the reference-rate reading recorded for the month the period starts in for ROR and DOR; fixed-rate types need none
    role: normative
    status: modeled
    span:
      start: 0
      end: 286
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:18.973Z'
id: REQ-valuation-lot-indices
type: req
---
A bond lot's interest periods after the first that have started must take the CPI reading of the month two months before the period starts for inflation-indexed types, and the reference-rate reading recorded for the month the period starts in for ROR and DOR; fixed-rate types need none.

## Context

Index readings are inputs of stored valuations (D-36); pl.CPIMonth fixes which CPI month sets a period. The reference-rate reading is stored per month it applies to, because picking the rate in force on the tenth business day before needs the statutory holiday calendar, which OpenFief does not have yet. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

docs/DESIGN.md D-36
