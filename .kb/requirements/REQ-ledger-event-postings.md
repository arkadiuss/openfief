---
title: Buys, sells, dividends with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Buys, sells, dividends with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.
logic_claims:
  - CLAIM-70A7256A397DB822
semantic_clauses:
  - Buys, sells, dividends with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b5fb46a5c51b844793a09e415bc0ec2b99f6f2a098cb503242af6b3efa0833cf
semantic_inventory:
  - claim_key: CLAIM-70A7256A397DB822
    claim_text: Buys, sells, dividends with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:51.843Z'
id: REQ-ledger-event-postings
type: req
---
Buys, sells, dividends with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.

## Context

Connectors map broker rows to these events, so the ledger offers one pure constructor per event that always produces a balanced transaction (D-19) and keeps every amount in its original currency (D-20); PLN conversion with NBP rates belongs to valuation. A fee may be in another currency than the trade, as IBKR charges commissions in the base currency. Fees are booked to the fee ledger account and are not added to a lot's cost in M0. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Every amount keeps its original currency. In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface

docs/DESIGN.md D-20

> Double-entry postings across cash, position, fee, withholding-tax and FX accounts.

docs/DESIGN.md D-19
