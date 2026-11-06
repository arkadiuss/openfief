---
title: The XTB connector must book each free-funds interest row of an xStation export as cash income, with the free-funds interest tax rows of its date as the tax withheld, and report a tax row with no single interest row of its date, or a refund, as an ingest problem.
status: open
text_ref: docs/DESIGN.md#D-152
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-152
  - lane:strict
semantic_text: The XTB connector must book each free-funds interest row of an xStation export as cash income, with the free-funds interest tax rows of its date as the tax withheld, and report a tax row with no single interest row of its date, or a refund, as an ingest problem.
logic_claims:
  - CLAIM-3148369A8B15BB1D
semantic_clauses:
  - The XTB connector must book each free-funds interest row of an xStation export as cash income, with the free-funds interest tax rows of its date as the tax withheld, and report a tax row with no single interest row of its date, or a refund, as an ingest problem.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: aa4542a1300e9e3937a8d426a632054408afa27568d1475a1fdab4707335eda6
semantic_inventory:
  - claim_key: CLAIM-3148369A8B15BB1D
    claim_text: The XTB connector must book each free-funds interest row of an xStation export as cash income, with the free-funds interest tax rows of its date as the tax withheld, and report a tax row with no single interest row of its date, or a refund, as an ingest problem
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
origin:
  kind: agent
  recorded_at: '2026-11-02T05:51:32.263Z'
id: REQ-ingest-xtb-free-funds-interest
type: req
---
The XTB connector must book each free-funds interest row of an xStation export as cash income, with the free-funds interest tax rows of its date as the tax withheld, and report a tax row with no single interest row of its date, or a refund, as an ingest problem.

## Context

D-152 names a broker's interest on uninvested cash as cash income booked with any tax withheld; XTB reports it as Free funds interest with a Free funds interest tax row of its own (R-24). Pairing the tax with the interest of its date is this implementation's reading.

## Source

> **[D-152]** The ledger has a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result, booked with any tax withheld. Performance counts it as return, not as money paid in.

docs/DESIGN.md#D-152
