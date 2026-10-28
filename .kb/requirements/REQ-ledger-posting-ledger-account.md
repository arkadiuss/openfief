---
title: A posting must go to a ledger account of a known kind in its own currency, and only a holding posting may carry a quantity of units.
status: open
text_ref: docs/DESIGN.md#D-19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-19
  - lane:strict
semantic_text: A posting must go to a ledger account of a known kind in its own currency, and only a holding posting may carry a quantity of units.
logic_claims:
  - CLAIM-927FCA7D4DE3FE5A
semantic_clauses:
  - A posting must go to a ledger account of a known kind in its own currency, and only a holding posting may carry a quantity of units.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e4067ed1135ab82eb716072ac19f924f270dd5d311abf8a2d32a4451852981f5
semantic_inventory:
  - claim_key: CLAIM-927FCA7D4DE3FE5A
    claim_text: A posting must go to a ledger account of a known kind in its own currency, and only a holding posting may carry a quantity of units
    role: normative
    status: modeled
    span:
      start: 0
      end: 131
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:36.973Z'
id: REQ-ledger-posting-ledger-account
type: req
---
A posting must go to a ledger account of a known kind in its own currency, and only a holding posting may carry a quantity of units.

## Context

D-19 records transactions as double-entry postings across cash, position, fee, withholding-tax and FX accounts. The ledger adds an equity ledger account per currency as the counterpart of deposits, withdrawals and opening lots, which otherwise could not balance; this is a modeling choice by Claude Code that the author has not confirmed. Keeping each posting in its ledger account's currency makes per-currency balances and balance assertions exact (D-20). Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Double-entry postings across cash, position, fee, withholding-tax and FX accounts.

docs/DESIGN.md D-19

> Every amount keeps its original currency.

docs/DESIGN.md D-20
