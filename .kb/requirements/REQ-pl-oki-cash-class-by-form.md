---
title: The Polish pack must classify a treasury savings bond as a savings asset, PLN cash on an OKI bank account or treasury bond account as a savings asset and on a brokerage account or fund register as an investment asset, PLN cash on an OKI account without a form and cash in another currency as other assets, and report that it can't classify a holding priced at market.
status: open
text_ref: docs/DESIGN.md#D-147
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-147
  - lane:strict
semantic_text: The Polish pack must classify a treasury savings bond as a savings asset, PLN cash on an OKI bank account or treasury bond account as a savings asset and on a brokerage account or fund register as an investment asset, PLN cash on an OKI account without a form and cash in another currency as other assets, and report that it can't classify a holding priced at market.
logic_claims:
  - CLAIM-91D35A378313F46B
semantic_clauses:
  - The Polish pack must classify a treasury savings bond as a savings asset, PLN cash on an OKI bank account or treasury bond account as a savings asset and on a brokerage account or fund register as an investment asset, PLN cash on an OKI account without a form and cash in another currency as other assets, and report that it can't classify a holding priced at market.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fc087e2dfcf2a445ad43fe149ef971f750b3c4e97136f0c3f10e0056e4fcf677
semantic_inventory:
  - claim_key: CLAIM-91D35A378313F46B
    claim_text: The Polish pack must classify a treasury savings bond as a savings asset, PLN cash on an OKI bank account or treasury bond account as a savings asset and on a brokerage account or fund register as an investment asset, PLN cash on an OKI account without a form and cash in another currency as other assets, and report that it can't classify a holding priced at market
    role: normative
    status: modeled
    span:
      start: 0
      end: 366
origin:
  kind: agent
  recorded_at: '2026-11-01T20:12:02.608Z'
id: REQ-pl-oki-cash-class-by-form
type: req
---
The Polish pack must classify a treasury savings bond as a savings asset, PLN cash on an OKI bank account or treasury bond account as a savings asset and on a brokerage account or fund register as an investment asset, PLN cash on an OKI account without a form and cash in another currency as other assets, and report that it can't classify a holding priced at market.

## Context

D-147 makes the account's form decide whether its PLN cash is a savings or an investment asset. Counting PLN cash on an account without a form, such as one created before forms were asked for, as other assets is a narrow reading that errs high, as D-132 does for instruments. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Why this changed

It supersedes REQ-pl-oki-exemption-class: D-147 makes the OKI account's form decide its PLN cash's class, replacing the narrow reading that PLN cash was savings only on an account that held treasury bonds in the year. The superseded requirement is closed.

## Source

> **[D-147]** The owner sets an OKI account's form when creating it: a bank account, a treasury bond account, a brokerage account or a fund register. The form decides whether its PLN cash is a savings or an investment asset.

docs/DESIGN.md#D-147
