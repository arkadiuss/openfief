---
title: OKI's tax exemptions must rise with CPI from 2030.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-127-6
  - lane:strict
text_ref: design-decisions:D-127#6
origin:
  kind: import
  ref: design-decisions:D-127#6
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: OKI's tax exemptions must rise with CPI from 2030.
logic_claims:
  - CLAIM-C51FCF1A5E0C362A
semantic_clauses:
  - OKI's tax exemptions must rise with CPI from 2030.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 32aadcecb7117532ca1233f4798448fb77cd888b636cef82b43ceed48b6e71f5
semantic_inventory:
  - claim_key: CLAIM-C51FCF1A5E0C362A
    claim_text: OKI's tax exemptions must rise with CPI from 2030
    role: normative
    status: modeled
    span:
      start: 0
      end: 49
id: REQ-pl-oki-exemptions-cpi-indexed
type: req
---
OKI's tax exemptions must rise with CPI from 2030.

## Context

The OKI act exempts up to 25,000 PLN of savings assets and up to 100,000 PLN of investment assets (REQ-pl-oki-tax-exemptions). The implementation raised the act's numbers as question #14 on docs/implementation-questions.md, and R-18 cites the published act (Dz.U. 2026 poz. 1098): the limits rise with CPI, rounded down to 100 PLN, from the 2030 tax year (art. 26 ust. 5, art. 44). The author recorded this as part of D-127 on 2026-10-31. OpenFief already reads CPI from GUS's final monthly values (D-104).

## Source

> OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account) are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030. Income inside OKI is outside PIT, so no Belka tax is due.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-127#6
