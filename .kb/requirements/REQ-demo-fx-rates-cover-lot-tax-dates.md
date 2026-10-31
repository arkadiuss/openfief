---
title: The demo loader must record NBP rates from before the tax rate date of the open date of the dataset's earliest foreign-currency lot, so that every demo holding has its net and liquidation values.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: The demo loader must record NBP rates from before the tax rate date of the open date of the dataset's earliest foreign-currency lot, so that every demo holding has its net and liquidation values.
logic_claims:
  - CLAIM-6918EAB86017CFB8
semantic_clauses:
  - The demo loader must record NBP rates from before the tax rate date of the open date of the dataset's earliest foreign-currency lot, so that every demo holding has its net and liquidation values.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6b3a82c8a1cd9e44298bcb394b6e75413798851372d841e0c2e8c5cbaaa58f82
semantic_inventory:
  - claim_key: CLAIM-6918EAB86017CFB8
    claim_text: The demo loader must record NBP rates from before the tax rate date of the open date of the dataset's earliest foreign-currency lot, so that every demo holding has its net and liquidation values
    role: normative
    status: modeled
    span:
      start: 0
      end: 194
origin:
  kind: agent
  recorded_at: '2026-10-31T12:17:20.434Z'
id: REQ-demo-fx-rates-cover-lot-tax-dates
type: req
---
The demo loader must record NBP rates from before the tax rate date of the open date of the dataset's earliest foreign-currency lot, so that every demo holding has its net and liquidation values.

## Context

A lot's cost converts to PLN at the NBP rate the Polish pack takes for tax on its open date, the business day before it (D-20), and net and liquidation values need that cost (D-77). The demo's IBKR opening lots were opened long before go-live while the loader's rates started two weeks before go-live, so every IBKR holding had a gross value but net and liquidation missing, and so did the account's and the overview's totals. Modeled by Claude Code for the demo-numbers task: the author reported that the demo's IBKR holdings showed net and liquidation values as missing for want of an FX rate, and that the price checks page showed no check; the author has not reviewed this requirement yet.

## Source

> The Polish jurisdiction pack always uses NBP rates for tax, whatever the UI uses.

docs/DESIGN.md D-20

> Every holding, account and total has three values, computed by the jurisdiction pack for each kind of product and tax wrapper

docs/DESIGN.md D-77
