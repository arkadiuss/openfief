---
title: An account's page must show a breadcrumb, the name with its sleeve, wrapper and sync pills, the owner's Upload statement action, its open items' banner, figure cards of gross with cash, net, liquidation and the time-weighted return since the charts' start, then the tabs Holdings, Transactions, Lots, Cash, Reconciliation and Account settings, each a link that works without JavaScript and an htmx swap, with every holding linked to its instrument.
status: open
text_ref: docs/DESIGN.md#D-195
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-195
  - lane:strict
semantic_text: An account's page must show a breadcrumb, the name with its sleeve, wrapper and sync pills, the owner's Upload statement action, its open items' banner, figure cards of gross with cash, net, liquidation and the time-weighted return since the charts' start, then the tabs Holdings, Transactions, Lots, Cash, Reconciliation and Account settings, each a link that works without JavaScript and an htmx swap, with every holding linked to its instrument.
logic_claims:
  - CLAIM-2014A39F59235DD2
semantic_clauses:
  - An account's page must show a breadcrumb, the name with its sleeve, wrapper and sync pills, the owner's Upload statement action, its open items' banner, figure cards of gross with cash, net, liquidation and the time-weighted return since the charts' start, then the tabs Holdings, Transactions, Lots, Cash, Reconciliation and Account settings, each a link that works without JavaScript and an htmx swap, with every holding linked to its instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5b9b66062cbc904dfcd12b9240b71e57e472bcdbda6b37e4e7b2332f4911c6e3
semantic_inventory:
  - claim_key: CLAIM-2014A39F59235DD2
    claim_text: An account's page must show a breadcrumb, the name with its sleeve, wrapper and sync pills, the owner's Upload statement action, its open items' banner, figure cards of gross with cash, net, liquidation and the time-weighted return since the charts' start, then the tabs Holdings, Transactions, Lots, Cash, Reconciliation and Account settings, each a link that works without JavaScript and an htmx swap, with every holding linked to its instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 447
origin:
  kind: agent
  recorded_at: '2026-11-06T10:43:28.616Z'
id: REQ-ui-account-page-tabs
type: req
---
An account's page must show a breadcrumb, the name with its sleeve, wrapper and sync pills, the owner's Upload statement action, its open items' banner, figure cards of gross with cash, net, liquidation and the time-weighted return since the charts' start, then the tabs Holdings, Transactions, Lots, Cash, Reconciliation and Account settings, each a link that works without JavaScript and an htmx swap, with every holding linked to its instrument.

## Context

Refines D-195 and D-196 with the design system's M0 account screen; the earlier content moves into the tabs. Modeled by Claude Code for the UI v2 pages slice; the author has not reviewed the wording.

## Source

> **[D-195]** Every figure links down to what it is made of: the total to its accounts, an account to its holdings, a holding to its instrument, an allocation bar to the holdings behind it. An instrument has its own page: its price history and source, its holdings and lots across accounts, its transactions, its price checks, its ticker and ISIN, and its OKI class.

docs/DESIGN.md#D-195

> **[D-196]** ... An item that one field or button settles ... is settled in place; any other links to its account or instrument. ... an account's page shows its own items at the top.

docs/DESIGN.md#D-196
