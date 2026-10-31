---
title: For OKI's yearly tax, values must count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-127-3
  - lane:strict
text_ref: design-decisions:D-127#3
origin:
  kind: import
  ref: design-decisions:D-127#3
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: For OKI's yearly tax, values must count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before.
logic_claims:
  - CLAIM-D1496BF061AAD8CB
semantic_clauses:
  - For OKI's yearly tax, values must count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: fc7a40006d0bda36ba1d737075f197997d7023cd5833198f4b7f96540de1842b
semantic_inventory:
  - claim_key: CLAIM-D1496BF061AAD8CB
    claim_text: For OKI's yearly tax, values must count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before
    role: normative
    status: modeled
    span:
      start: 0
      end: 238
id: REQ-pl-oki-tax-base-valuation
type: req
---
For OKI's yearly tax, values must count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before.

## Why this changed

REQ-pl-oki-yearly-tax, from D-76, computed OKI's yearly tax from OpenFief's stored daily valuations. The published act (Dz.U. 2026 poz. 1098), which the author recorded as D-127 on 2026-10-31, sets how each asset is valued for this tax, and its foreign-currency rule, the NBP average rate of the last business day before, need not match the rate a stored valuation uses, so this requirement replaces that part of the old one.

## Context

The implementation raised OKI's act as question #14 on docs/implementation-questions.md; it averaged gross values with cash from the stored valuations (REQ-pl-oki-act-as-data and REQ-views-yearly-taxes, which the author has not reviewed). R-18 cites the act's valuation rule (art. 24): listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, foreign amounts at the NBP average rate of the last business day before, and a day without a valuation takes the last one. Polish income tax converts foreign income the same way (REQ-pl-tax-fx-rate-date).

## Source

> OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account) are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030. Income inside OKI is outside PIT, so no Belka tax is due.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-127#3
