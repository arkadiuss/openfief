---
title: The XTB connector must book each closed CFD position's Profit/Loss as cash income on the UTC date it closed, keyed by its position ID, and report every open CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
status: open
text_ref: docs/DESIGN.md#D-152
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-152
  - lane:strict
semantic_text: The XTB connector must book each closed CFD position's Profit/Loss as cash income on the UTC date it closed, keyed by its position ID, and report every open CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
logic_claims:
  - CLAIM-3B0BD8DCE3FEC389
semantic_clauses:
  - The XTB connector must book each closed CFD position's Profit/Loss as cash income on the UTC date it closed, keyed by its position ID, and report every open CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 87f46c62f4f3353dca687944d6ad893a5da7490197404528089ae4a33619d149
semantic_inventory:
  - claim_key: CLAIM-3B0BD8DCE3FEC389
    claim_text: The XTB connector must book each closed CFD position's Profit/Loss as cash income on the UTC date it closed, keyed by its position ID, and report every open CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-11-02T05:51:39.373Z'
id: REQ-ingest-xtb-cfd-closed-cash-income
type: req
---
The XTB connector must book each closed CFD position's Profit/Loss as cash income on the UTC date it closed, keyed by its position ID, and report every open CFD position, and every cash operation of a category other than a stock or ETF, as an unsupported ingest problem.

## Context

D-24 imports closed CFD results as cash profit or loss so cash reconciles, and D-152 books them as cash income. Reading the result from the Closed Positions sheet's Profit/Loss column, in the account currency, is this implementation's reading; whether that column is net of commission and swap, and whether Cash Operations repeats it, waits on a real CFD export (implementation question 81).

## Why this changed

The author decided D-152, which books a closed CFD's result as cash income, so it replaces REQ-ingest-xtb-cfd-flagged: that requirement reported closed CFD positions as unsupported, which left the account's cash unreconciled, and is closed.

## Source

> **[D-152]** The ledger has a cash income kind for income that no holding produces, such as a broker's interest on uninvested cash or a closed CFD's result, booked with any tax withheld. Performance counts it as return, not as money paid in.

docs/DESIGN.md#D-152

> **[D-24]** XTB in M0 covers real stocks and ETFs only. Closed CFD results are imported as cash profit or loss so cash reconciles; open CFD positions are flagged as unsupported.

docs/DESIGN.md#D-24
