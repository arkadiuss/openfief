---
title: 'The XTB connector must map an uploaded xStation export of one account in transactions mode: deposits and withdrawals as transactions, stock and ETF buys, sells and dividends with their withheld tax as ticker entries named by the XTB account, with amounts rounded to the account currency''s minor unit.'
status: open
text_ref: docs/research.md#R-24
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-research-md-r-24
  - lane:strict
semantic_text: 'The XTB connector must map an uploaded xStation export of one account in transactions mode: deposits and withdrawals as transactions, stock and ETF buys, sells and dividends with their withheld tax as ticker entries named by the XTB account, with amounts rounded to the account currency''s minor unit.'
logic_claims:
  - CLAIM-37023577BE70B9C6
semantic_clauses:
  - 'The XTB connector must map an uploaded xStation export of one account in transactions mode: deposits and withdrawals as transactions, stock and ETF buys, sells and dividends with their withheld tax as ticker entries named by the XTB account, with amounts rounded to the account currency''s minor unit.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 03a1cc25f128f349c585ab19473e9a08e3026d02cbaf8c12e636bcd970f19816
semantic_inventory:
  - claim_key: CLAIM-37023577BE70B9C6
    claim_text: 'The XTB connector must map an uploaded xStation export of one account in transactions mode: deposits and withdrawals as transactions, stock and ETF buys, sells and dividends with their withheld tax as ticker entries named by the XTB account, with amounts rounded to the account currency''s minor unit'
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-10-31T12:55:32.983Z'
id: REQ-ingest-xtb-export-mapping
type: req
---
The XTB connector must map an uploaded xStation export of one account in transactions mode: deposits and withdrawals as transactions, stock and ETF buys, sells and dividends with their withheld tax as ticker entries named by the XTB account, with amounts rounded to the account currency's minor unit.

## Context

D-9 brings XTB in through uploaded xStation exports, and R-24 records the export's layout: one account per file, English labels, binary float amounts. D-72 names entries by the broker account. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's xStation "Historia operacji" export ... holds one account. Its labels are English, and it has three sheets, Closed Positions, Cash Operations and Open Positions.

docs/research.md#R-24
