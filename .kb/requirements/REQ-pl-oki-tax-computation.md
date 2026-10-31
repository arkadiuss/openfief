---
title: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.
status: open
text_ref: docs/DESIGN.md#D-127
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-127
  - lane:strict
semantic_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.
logic_claims:
  - CLAIM-4B7BFD6A08E3D1FC
semantic_clauses:
  - The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 93a6e1a569a54bb74fe8016f9634529329b72bbefe916a4b126accee4e5392cf
semantic_inventory:
  - claim_key: CLAIM-4B7BFD6A08E3D1FC
    claim_text: The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known
    role: exception
    status: modeled
    span:
      start: 0
      end: 440
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:47.300Z'
id: REQ-pl-oki-tax-computation
type: req
---
The Polish pack must compute OKI's yearly tax from each OKI account's days held, averaging each class's end-of-day values plus its share of the payments counted up to each day's withdrawals over the days held, half up to grosze, exempting savings assets first then investment assets up to the year's amounts, and taxing the rest at the year's rate half up to grosze, with no rate or tax while the year's rate or exempt amounts are not known.

## Context

D-127 fixes the act's base, exemptions and rates; the split of payments between classes follows art. 26 ust. 3 of the act (R-18). Rounding is a narrow reading listed in docs/implementation-questions.md. Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-pl-oki-act-as-data: D-127 replaces D-76's balance limit with the act's exemptions and yearly rates, in force from 2027-01-01. The superseded requirement is closed.

## Source

> **[D-127]** OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets ... and up to 100,000 PLN of investment assets ... are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030.

docs/DESIGN.md#D-127
