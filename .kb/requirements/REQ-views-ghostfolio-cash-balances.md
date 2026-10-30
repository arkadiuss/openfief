---
title: The Ghostfolio view must give each owned account, in the Ghostfolio user's base currency, one cash balance at the end of every date on which the account's cash in that currency changes, counting the transactions it leaves out, must leave out cash in any other currency as a problem per account and currency, and must leave out an account's balances when one is too large for a double.
status: open
text_ref: docs/DESIGN.md#D-108
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-108
  - lane:strict
semantic_text: The Ghostfolio view must give each owned account, in the Ghostfolio user's base currency, one cash balance at the end of every date on which the account's cash in that currency changes, counting the transactions it leaves out, must leave out cash in any other currency as a problem per account and currency, and must leave out an account's balances when one is too large for a double.
logic_claims:
  - CLAIM-14CEF421E9EEFB04
semantic_clauses:
  - The Ghostfolio view must give each owned account, in the Ghostfolio user's base currency, one cash balance at the end of every date on which the account's cash in that currency changes, counting the transactions it leaves out, must leave out cash in any other currency as a problem per account and currency, and must leave out an account's balances when one is too large for a double.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a80f6b08b7eb9bcca006914b7f8eebd45cb03030989a63c061511dd86a4de533
semantic_inventory:
  - claim_key: CLAIM-14CEF421E9EEFB04
    claim_text: The Ghostfolio view must give each owned account, in the Ghostfolio user's base currency, one cash balance at the end of every date on which the account's cash in that currency changes, counting the transactions it leaves out, must leave out cash in any other currency as a problem per account and currency, and must leave out an account's balances when one is too large for a double
    role: normative
    status: modeled
    span:
      start: 0
      end: 383
origin:
  kind: agent
  recorded_at: '2026-10-30T01:21:08.079Z'
id: REQ-views-ghostfolio-cash-balances
type: req
---
The Ghostfolio view must give each owned account, in the Ghostfolio user's base currency, one cash balance at the end of every date on which the account's cash in that currency changes, counting the transactions it leaves out, must leave out cash in any other currency as a problem per account and currency, and must leave out an account's balances when one is too large for a double.

## Context

Refines the firm REQ-views-ghostfolio-cash-balance (D-108). A Ghostfolio account has one balance in its own currency and the view creates its accounts in the user's base currency, so cash in another currency can't be shown: the narrowest reading leaves it out and reports it (docs/implementation-questions.md, Ghostfolio view). Ghostfolio keeps one balance per account and date, so a balance at each change carries forward in its charts. Modeled by Claude Code for the Ghostfolio work package (impl/ghostfolio-2); the author has not reviewed it yet.

## Source

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

docs/DESIGN.md#D-108
