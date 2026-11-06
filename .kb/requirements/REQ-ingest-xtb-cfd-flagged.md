---
title: The XTB connector must report every open or closed CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
status: closed
text_ref: docs/DESIGN.md#D-24
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-24
  - lane:strict
semantic_text: The XTB connector must report every open or closed CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
logic_claims:
  - CLAIM-3DCDA3C51601722A
semantic_clauses:
  - The XTB connector must report every open or closed CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0977ecacb180af9eddf4afd869a7e8509b62cb43e6247eca73ae88d76d807045
semantic_inventory:
  - claim_key: CLAIM-3DCDA3C51601722A
    claim_text: The XTB connector must report every open or closed CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem
    role: normative
    status: modeled
    span:
      start: 0
      end: 163
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:43.304Z'
id: REQ-ingest-xtb-cfd-flagged
type: req
---
The XTB connector must report every open or closed CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.

## Context

D-24 limits XTB to stocks and ETFs and flags open CFDs. Booking closed CFD results as cash needs a ledger kind that doesn't exist yet, so they are flagged too until the author decides (implementation questions). Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB in M0 covers real stocks and ETFs only. Closed CFD results are imported as cash profit or loss so cash reconciles; open CFD positions are flagged as unsupported.

docs/DESIGN.md#D-24
