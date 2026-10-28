---
title: In opening mode, each lot-level IBKR Flex open position of a stock or ETF must become an opening lot with its open date, units and cost basis, recorded on the statement's last day, and each currency's ending cash must become the account's opening cash.
status: open
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
semantic_text: In opening mode, each lot-level IBKR Flex open position of a stock or ETF must become an opening lot with its open date, units and cost basis, recorded on the statement's last day, and each currency's ending cash must become the account's opening cash.
logic_claims:
  - CLAIM-9D55699154DCDD5D
semantic_clauses:
  - In opening mode, each lot-level IBKR Flex open position of a stock or ETF must become an opening lot with its open date, units and cost basis, recorded on the statement's last day, and each currency's ending cash must become the account's opening cash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fe0db19aaea4c99abac6ee4bff4ec746d8a0f8494af47e0d55c399d2772a0a14
semantic_inventory:
  - claim_key: CLAIM-9D55699154DCDD5D
    claim_text: In opening mode, each lot-level IBKR Flex open position of a stock or ETF must become an opening lot with its open date, units and cost basis, recorded on the statement's last day, and each currency's ending cash must become the account's opening cash
    role: normative
    status: modeled
    span:
      start: 0
      end: 251
origin:
  kind: agent
  recorded_at: '2026-10-28T06:23:39.692Z'
id: REQ-ibkrflex-opening
type: req
---
In opening mode, each lot-level IBKR Flex open position of a stock or ETF must become an opening lot with its open date, units and cost basis, recorded on the statement's last day, and each currency's ending cash must become the account's opening cash.

## Context

D-23 seeds accounts from the broker's current lot-level positions instead of full history, so FIFO lots and their PLN cost stay correct from go-live. Opening cash is booked as a deposit (or a withdrawal for a negative balance) on the go-live date so that cash balance assertions hold from then on; there is no dedicated opening-cash transaction kind yet. Refines REQ-AUTO-3C2A493C661181C7. Modeled by Claude Code for the M0 IBKR Flex connector task; the author has not reviewed it yet. The Flex schema is IBKR's public Flex Query reference, not yet verified against a real account (OQ-22).

## Source

> Accounts are seeded from each broker's current lot-level positions (opening lots), with transactions from go-live. Full history backfill arrives with the tax pack.

docs/DESIGN.md D-23
