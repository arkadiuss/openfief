---
title: In transactions mode, IBKR Flex stock and ETF executions must become buys and sells keyed by tradeID, currency-pair executions FX conversions, and deposits, withdrawals, other fees and dividends transactions keyed by transactionID; in opening mode those rows must be skipped.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: In transactions mode, IBKR Flex stock and ETF executions must become buys and sells keyed by tradeID, currency-pair executions FX conversions, and deposits, withdrawals, other fees and dividends transactions keyed by transactionID; in opening mode those rows must be skipped.
logic_claims:
  - CLAIM-DEEB2ADF98D7E880
semantic_clauses:
  - In transactions mode, IBKR Flex stock and ETF executions must become buys and sells keyed by tradeID, currency-pair executions FX conversions, and deposits, withdrawals, other fees and dividends transactions keyed by transactionID; in opening mode those rows must be skipped.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5a18a670f00aa3a98adbfe381d0089724d26421bdae292cd090dcbe662db5de6
semantic_inventory:
  - claim_key: CLAIM-DEEB2ADF98D7E880
    claim_text: In transactions mode, IBKR Flex stock and ETF executions must become buys and sells keyed by tradeID, currency-pair executions FX conversions, and deposits, withdrawals, other fees and dividends transactions keyed by transactionID; in opening mode those rows must be skipped
    role: normative
    status: modeled
    span:
      start: 0
      end: 274
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:47.365Z'
id: REQ-ibkrflex-transactions
type: req
---
In transactions mode, IBKR Flex stock and ETF executions must become buys and sells keyed by tradeID, currency-pair executions FX conversions, and deposits, withdrawals, other fees and dividends transactions keyed by transactionID; in opening mode those rows must be skipped.

## Context

D-23 records transactions from go-live, and D-18 keys re-imports by source plus the institution's own ID, which for IBKR is the tradeID or transactionID. Dates are taken as calendar dates as IBKR reports them, without time zone conversion. Rows dated before go-live are the opening lots' history, so an opening statement books none of them. Refines REQ-AUTO-34242A3F858F647D and REQ-AUTO-FD618A602F7F9106. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The Flex schema is IBKR's public Flex Query reference, not yet verified against a real account (OQ-22).

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live. Full history backfill arrives with the tax pack.

docs/DESIGN.md D-23
