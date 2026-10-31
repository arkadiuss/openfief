---
title: OKI's yearly tax rate must be 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-127-5
  - lane:strict
text_ref: design-decisions:D-127#5
origin:
  kind: import
  ref: design-decisions:D-127#5
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: OKI's yearly tax rate must be 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places.
logic_claims:
  - CLAIM-60476A8AC93112A6
semantic_clauses:
  - OKI's yearly tax rate must be 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 43dfeceb9b436f45654fea095ae8c0524f632b9e396fada1474cf121683deb47
semantic_inventory:
  - claim_key: CLAIM-60476A8AC93112A6
    claim_text: OKI's yearly tax rate must be 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places
    role: normative
    status: modeled
    span:
      start: 0
      end: 173
id: REQ-pl-oki-tax-rate
type: req
---
OKI's yearly tax rate must be 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places.

## Why this changed

REQ-pl-oki-yearly-tax, from D-76, took OKI's tax rate from the act without stating it, because sources disagreed on the rate while the act was a draft (R-18). The published act (Dz.U. 2026 poz. 1098), which the author recorded as D-127 on 2026-10-31, sets the rate as a fixed 0.85% for 2027 and a formula on the NBP reference rate after that, so this requirement replaces that part of the old one.

## Context

The implementation holds the act's rate as data, with no act in force (REQ-pl-oki-act-as-data, which the author has not reviewed), and raised the rate as question #14 on docs/implementation-questions.md. R-18 cites the act: 0.85% for 2027 (art. 43), then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places and announced by 30 November (art. 25). The treasury bond connector already delivers the NBP reference rate (D-94, D-102). R-18 notes the act doesn't state how the tax amount itself is rounded.

## Source

> OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account) are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030. Income inside OKI is outside PIT, so no Belka tax is due.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-127#5
