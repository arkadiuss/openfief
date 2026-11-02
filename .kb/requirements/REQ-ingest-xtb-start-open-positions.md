---
title: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position quoted in the account''s currency, on the report''s date, first acquired on its open date at its volume times its open price, by ticker until the owner confirms its ISIN.'
status: closed
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
semantic_text: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position quoted in the account''s currency, on the report''s date, first acquired on its open date at its volume times its open price, by ticker until the owner confirms its ISIN.'
logic_claims:
  - CLAIM-80FF230DDDA849F0
semantic_clauses:
  - 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position quoted in the account''s currency, on the report''s date, first acquired on its open date at its volume times its open price, by ticker until the owner confirms its ISIN.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 36a2080efa0689f2ecb19b524fe35d98325f473976cf95543cc422af998751c6
semantic_inventory:
  - claim_key: CLAIM-80FF230DDDA849F0
    claim_text: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position quoted in the account''s currency, on the report''s date, first acquired on its open date at its volume times its open price, by ticker until the owner confirms its ISIN'
    role: normative
    status: modeled
    span:
      start: 0
      end: 301
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:23.188Z'
id: REQ-ingest-xtb-start-open-positions
type: req
---
The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position quoted in the account's currency, on the report's date, first acquired on its open date at its volume times its open price, by ticker until the owner confirms its ISIN.

## Context

D-136 starts an XTB account from an export of its whole history or from a current export's open positions, one opening lot per position at its open time and price. R-24 gives open prices in the instrument's currency with no rate, so a position quoted in another currency is left out as a problem until the author decides (implementation question 52). It replaces the earlier refusal of opening batches. Recorded by the XTB start implementation agent (impl/xtb-start) for D-136; default firmness, not reviewed by the author yet; firm requirements for D-136 are to come from the design session.

## Why this changed

D-136, decided by the author, replaces the reading REQ-ingest-xtb-opening-refused recorded: an XTB account now starts from its open positions too, and later statements continue it from go-live, so the earlier narrow reading no longer holds.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

docs/DESIGN.md#D-136
