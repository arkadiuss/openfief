---
title: OKI's yearly tax must exempt up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account), at most 100,000 PLN together with savings assets first, and must exempt no other assets.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-127-4
  - lane:strict
text_ref: design-decisions:D-127#4
origin:
  kind: import
  ref: design-decisions:D-127#4
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: OKI's yearly tax must exempt up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account), at most 100,000 PLN together with savings assets first, and must exempt no other assets.
logic_claims:
  - CLAIM-96003E0598305BA2
semantic_clauses:
  - OKI's yearly tax must exempt up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account), at most 100,000 PLN together with savings assets first, and must exempt no other assets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 1693ca08ad504d67841c4d29d8c8cf2551354463baface523c406a86fdbe4b14
semantic_inventory:
  - claim_key: CLAIM-96003E0598305BA2
    claim_text: OKI's yearly tax must exempt up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account), at most 100,000 PLN together with savings assets first, and must exempt no other assets
    role: exception
    status: modeled
    span:
      start: 0
      end: 392
id: REQ-pl-oki-tax-exemptions
type: req
---
OKI's yearly tax must exempt up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account), at most 100,000 PLN together with savings assets first, and must exempt no other assets.

## Why this changed

REQ-pl-oki-yearly-tax, from D-76, taxed the part of the average balance above one statutory limit, taken from the act once known; R-18's drafts had a single 100,000 PLN limit on qualified assets. The published act (Dz.U. 2026 poz. 1098), which the author recorded as D-127 on 2026-10-31, has no single limit: it exempts savings assets and investment assets up to separate amounts, at most 100,000 PLN together, and other assets not at all, so this requirement replaces that part of the old one.

## Context

The implementation raised OKI's act, its limit among its numbers, as question #14 on docs/implementation-questions.md, and its narrow reading taxes the part of the average daily balance above one limit (REQ-pl-oki-act-as-data, which the author has not reviewed). R-18 cites the act's exemptions (art. 26 ust. 1, 2 and 4): savings assets also include the lowest-risk PRIIP funds and PLN cash on a bond account, and a fund qualifies as an investment asset when at least 70% of its assets are such Polish assets. REQ-pl-oki-exemptions-cpi-indexed says how the amounts change from 2030.

## Source

> OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account) are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030. Income inside OKI is outside PIT, so no Belka tax is due.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-127#4
