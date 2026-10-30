---
title: The built-in UI's overview must show the portfolio's gross, net and liquidation values in the reporting currency, each account's three values and valuation status, the allocation by asset class, currency, institution or sleeve, a value-over-time chart of money with cash from go-live, and a separate performance chart of the time-weighted return index from go-live, 100 on go-live, from the read API, without computing performance itself.
status: open
text_ref: docs/DESIGN.md#D-106
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-106
  - lane:strict
semantic_text: The built-in UI's overview must show the portfolio's gross, net and liquidation values in the reporting currency, each account's three values and valuation status, the allocation by asset class, currency, institution or sleeve, a value-over-time chart of money with cash from go-live, and a separate performance chart of the time-weighted return index from go-live, 100 on go-live, from the read API, without computing performance itself.
logic_claims:
  - CLAIM-3775A4117F179150
semantic_clauses:
  - The built-in UI's overview must show the portfolio's gross, net and liquidation values in the reporting currency, each account's three values and valuation status, the allocation by asset class, currency, institution or sleeve, a value-over-time chart of money with cash from go-live, and a separate performance chart of the time-weighted return index from go-live, 100 on go-live, from the read API, without computing performance itself.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8e848bf1db6ba5ac233e48488d7302803fbbce9be4371309cd8ca5ad7f61b031
semantic_inventory:
  - claim_key: CLAIM-3775A4117F179150
    claim_text: The built-in UI's overview must show the portfolio's gross, net and liquidation values in the reporting currency, each account's three values and valuation status, the allocation by asset class, currency, institution or sleeve, a value-over-time chart of money with cash from go-live, and a separate performance chart of the time-weighted return index from go-live, 100 on go-live, from the read API, without computing performance itself
    role: normative
    status: modeled
    span:
      start: 0
      end: 437
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:54.637Z'
id: REQ-ui-overview-values
type: req
---
The built-in UI's overview must show the portfolio's gross, net and liquidation values in the reporting currency, each account's three values and valuation status, the allocation by asset class, currency, institution or sleeve, a value-over-time chart of money with cash from go-live, and a separate performance chart of the time-weighted return index from go-live, 100 on go-live, from the read API, without computing performance itself.

## Context

Supersedes REQ-ui-overview, which drew the time-weighted return as the value-over-time chart: D-106 makes value and performance separate charts, and D-77 gives totals three values. Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-ui-overview: D-106 separates the value and performance charts; D-77 gives three values. The superseded requirement is closed.

## Source

> **[D-106]** Value and performance are separate charts. The value-over-time chart shows money, cash included, with D-77's switch between gross, net and liquidation. The performance chart shows the time-weighted return (D-61) as an index from 100 at go-live, computed on gross value, next to the model portfolio (D-53). Valuation totals include cash.

docs/DESIGN.md#D-106
