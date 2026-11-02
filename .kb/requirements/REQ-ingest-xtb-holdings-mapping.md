---
title: The XTB connector must report, from an xStation export's Open Positions in either mode, each stock or ETF ticker's summed open volume at the end of the report's date, and report a ticker one of whose positions it can't read as an ingest problem instead of asserting it.
status: open
text_ref: docs/DESIGN.md#D-155
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-155
  - lane:strict
semantic_text: The XTB connector must report, from an xStation export's Open Positions in either mode, each stock or ETF ticker's summed open volume at the end of the report's date, and report a ticker one of whose positions it can't read as an ingest problem instead of asserting it.
logic_claims:
  - CLAIM-6FBC3F20F446E60C
semantic_clauses:
  - The XTB connector must report, from an xStation export's Open Positions in either mode, each stock or ETF ticker's summed open volume at the end of the report's date, and report a ticker one of whose positions it can't read as an ingest problem instead of asserting it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1d31528b3f874db1440dd41826e59c45155df454d7ed164f5155baa97cf24509
semantic_inventory:
  - claim_key: CLAIM-6FBC3F20F446E60C
    claim_text: The XTB connector must report, from an xStation export's Open Positions in either mode, each stock or ETF ticker's summed open volume at the end of the report's date, and report a ticker one of whose positions it can't read as an ingest problem instead of asserting it
    role: normative
    status: modeled
    span:
      start: 0
      end: 268
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:42.854Z'
id: REQ-ingest-xtb-holdings-mapping
type: req
---
The XTB connector must report, from an xStation export's Open Positions in either mode, each stock or ETF ticker's summed open volume at the end of the report's date, and report a ticker one of whose positions it can't read as an ingest problem instead of asserting it.

## Context

D-155 has Open Positions assert the account's holdings; asserting in both modes, skipping exports without a report date and refusing a partial sum are this implementation's reading. A ticker without open positions is not asserted (implementation question 82).

## Source

> **[D-155]** An XTB export's `Open Positions` sheet also asserts the account's holdings (D-18): for each ticker with a confirmed ISIN (D-116), the summed volume of its open positions on the report's date. A ticker not confirmed yet is asserted once the owner confirms it.

docs/DESIGN.md#D-155
