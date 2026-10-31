---
title: Income inside OKI must be outside PIT, so no Belka tax is due on it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-127-7
  - lane:strict
text_ref: design-decisions:D-127#7
origin:
  kind: import
  ref: design-decisions:D-127#7
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: Income inside OKI must be outside PIT, so no Belka tax is due on it.
logic_claims:
  - CLAIM-FF92B641709AB459
semantic_clauses:
  - Income inside OKI must be outside PIT, so no Belka tax is due on it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 37f9643f24c3e5782b4e7611a163169e40e53ccb66b5706792a5893c87394c8e
semantic_inventory:
  - claim_key: CLAIM-FF92B641709AB459
    claim_text: Income inside OKI must be outside PIT, so no Belka tax is due on it
    role: normative
    status: modeled
    span:
      start: 0
      end: 67
id: REQ-pl-oki-income-outside-pit
type: req
---
Income inside OKI must be outside PIT, so no Belka tax is due on it.

## Context

D-76 already says no Belka tax is due inside OKI (REQ-pl-oki-no-belka-tax), from R-18's reading of the draft act. The implementation raised OKI's act as question #14 on docs/implementation-questions.md, and R-18 now cites the published act (Dz.U. 2026 poz. 1098): income from OKI assets is outside the PIT act (art. 32, adding art. 2 ust. 1 pkt 10 to it), so no Belka tax is due inside OKI, and a loss inside OKI is no tax loss (art. 5 ust. 1 pkt 3 lit. c). The author restated the rule with that reason in D-127 on 2026-10-31; the yearly tax on asset value applies instead (REQ-pl-oki-act-in-force-2027).

## Source

> OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account) are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030. Income inside OKI is outside PIT, so no Belka tax is due.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-127#7
