---
title: 'The Ghostfolio view must give each owned account one cash balance at the end of every date on which the account''s cash in that owned account''s currency changes, counting the transactions it leaves out: the account''s own in the Ghostfolio user''s base currency, and each cash account in its currency, which holds only that balance; and must leave out an owned account''s balances when one is too large for a double.'
status: open
text_ref: docs/DESIGN.md#D-133
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-133
  - lane:strict
semantic_text: 'The Ghostfolio view must give each owned account one cash balance at the end of every date on which the account''s cash in that owned account''s currency changes, counting the transactions it leaves out: the account''s own in the Ghostfolio user''s base currency, and each cash account in its currency, which holds only that balance; and must leave out an owned account''s balances when one is too large for a double.'
logic_claims:
  - CLAIM-ADE19139887CAFDE
semantic_clauses:
  - 'The Ghostfolio view must give each owned account one cash balance at the end of every date on which the account''s cash in that owned account''s currency changes, counting the transactions it leaves out: the account''s own in the Ghostfolio user''s base currency, and each cash account in its currency, which holds only that balance; and must leave out an owned account''s balances when one is too large for a double.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d88b894f8187c4e248586c62ae4a887957e525cda929e927a7e9d5276d9c49a2
semantic_inventory:
  - claim_key: CLAIM-ADE19139887CAFDE
    claim_text: 'The Ghostfolio view must give each owned account one cash balance at the end of every date on which the account''s cash in that owned account''s currency changes, counting the transactions it leaves out: the account''s own in the Ghostfolio user''s base currency, and each cash account in its currency, which holds only that balance; and must leave out an owned account''s balances when one is too large for a double'
    role: normative
    status: modeled
    span:
      start: 0
      end: 411
origin:
  kind: agent
  recorded_at: '2026-10-31T19:00:29.572Z'
id: REQ-views-ghostfolio-cash-balances-per-currency
type: req
---
The Ghostfolio view must give each owned account one cash balance at the end of every date on which the account's cash in that owned account's currency changes, counting the transactions it leaves out: the account's own in the Ghostfolio user's base currency, and each cash account in its currency, which holds only that balance; and must leave out an owned account's balances when one is too large for a double.

## Context

D-108 has each owned account's cash balance follow OpenFief's cash, so Ghostfolio's totals include it; D-133 puts each other currency of an account's cash in an owned account of its own, holding only that balance. Refines REQ-views-ghostfolio-cash-balance (firm, D-108). Modeled by Claude Code while building D-133; the author has not reviewed this wording.

## Why this changed

REQ-views-ghostfolio-cash-balances left an account's cash in any currency other than the Ghostfolio user's base currency out as a problem, so Ghostfolio's totals missed it; D-133 decided that it goes into an owned account per currency instead.

## Source

> Each currency of an account's cash other than its Ghostfolio account's gets an owned Ghostfolio account of its own, holding only that balance, with a UUID derived from the instance, the account and the currency, so Ghostfolio's totals include all of the account's cash.

docs/DESIGN.md#D-133
