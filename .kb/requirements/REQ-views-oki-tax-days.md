---
title: The read API's yearly tax must build each OKI account's days from its first transaction through the year's end or the earliest valued-through date, with valued holdings and cash per exemption class at the NBP rate of the day before and the day's deposits and withdrawals, and list the instruments counted without a class.
status: open
text_ref: docs/DESIGN.md#D-127
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-127
  - lane:strict
semantic_text: The read API's yearly tax must build each OKI account's days from its first transaction through the year's end or the earliest valued-through date, with valued holdings and cash per exemption class at the NBP rate of the day before and the day's deposits and withdrawals, and list the instruments counted without a class.
logic_claims:
  - CLAIM-78B2EF1DF93B3B2B
semantic_clauses:
  - The read API's yearly tax must build each OKI account's days from its first transaction through the year's end or the earliest valued-through date, with valued holdings and cash per exemption class at the NBP rate of the day before and the day's deposits and withdrawals, and list the instruments counted without a class.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bbc6210b114ad36f3c21c3999f147c92ed3095368fc592b11714e4f2bdd4d5d2
semantic_inventory:
  - claim_key: CLAIM-78B2EF1DF93B3B2B
    claim_text: The read API's yearly tax must build each OKI account's days from its first transaction through the year's end or the earliest valued-through date, with valued holdings and cash per exemption class at the NBP rate of the day before and the day's deposits and withdrawals, and list the instruments counted without a class
    role: normative
    status: modeled
    span:
      start: 0
      end: 320
origin:
  kind: agent
  recorded_at: '2026-10-31T18:34:56.741Z'
id: REQ-views-oki-tax-days
type: req
---
The read API's yearly tax must build each OKI account's days from its first transaction through the year's end or the earliest valued-through date, with valued holdings and cash per exemption class at the NBP rate of the day before and the day's deposits and withdrawals, and list the instruments counted without a class.

## Context

D-127 computes the tax from stored daily valuations; D-132 lists what the owner should classify. The held-from date and a year in progress are narrow readings (docs/implementation-questions.md). Modeled by Claude Code for the taxes work package; the author has not reviewed it yet.

## Why this changed

It replaces REQ-views-yearly-taxes: D-127 computes OKI's tax from each account's days per exemption class, not from summed daily gross totals. The superseded requirement is closed.

## Source

> **[D-127]** OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets ... and up to 100,000 PLN of investment assets ... are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030.

docs/DESIGN.md#D-127
