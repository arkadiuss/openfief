---
title: OKI's yearly tax must follow the OKI act in force from 2027-01-01.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-127-1
  - lane:strict
text_ref: design-decisions:D-127#1
origin:
  kind: import
  ref: design-decisions:D-127#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: OKI's yearly tax must follow the OKI act in force from 2027-01-01.
logic_claims:
  - CLAIM-E602583E1CCB78AB
semantic_clauses:
  - OKI's yearly tax must follow the OKI act in force from 2027-01-01.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: c00ba4acff5799d2577017ad4f7ad16fe0d30a05bb70066df42f08411122eaa7
semantic_inventory:
  - claim_key: CLAIM-E602583E1CCB78AB
    claim_text: OKI's yearly tax must follow the OKI act in force from 2027-01-01
    role: normative
    status: modeled
    span:
      start: 0
      end: 65
id: REQ-pl-oki-act-in-force-2027
type: req
---
OKI's yearly tax must follow the OKI act in force from 2027-01-01.

## Why this changed

REQ-pl-oki-yearly-tax, from D-76, computed OKI's yearly tax only once the act was in force, with the limit and rate taken from the act, because neither the act's status nor its numbers were known when the author decided it. The act is now published (Dz.U. 2026 poz. 1098), and on 2026-10-31 the author recorded its rules as D-127, starting with its in-force date of 2027-01-01. The old requirement is split into this one, REQ-pl-oki-tax-base-average-value, REQ-pl-oki-tax-base-valuation, REQ-pl-oki-tax-exemptions and REQ-pl-oki-tax-rate.

## Context

D-76 gives every account a tax wrapper, OKI among the Polish pack's (REQ-pl-tax-wrappers). The implementation holds the act as data with no in-force date, so it never computes the tax and the read API says so (REQ-pl-oki-act-as-data, which the author has not reviewed), and it raised the act's in-force date, limit and rate as question #14 on docs/implementation-questions.md. R-18 now cites the published act (ustawa z dnia 3 lipca 2026 r. o osobistych kontach inwestycyjnych), which is in force from 2027-01-01 (art. 46), and the author encoded it as D-127.

## Source

> OKI's yearly tax follows the act in force from 2027-01-01. Its base is the sum over the owner's OKI accounts of each account's average value: the sum of its end-of-day values over the tax year, plus each day's payments in up to that day's withdrawals, divided by the days of the year the account was held. Values count listed securities at market price, treasury savings bonds at nominal plus accrued indexation or capitalization, and cash, with foreign amounts at the NBP average rate of the last business day before. Up to 25,000 PLN of the average of savings assets (PLN cash, treasury savings bonds, PLN Treasury bills, the lowest-risk funds) and up to 100,000 PLN of investment assets (shares of companies with capital in PLN, PLN bonds, qualifying funds, PLN cash on a brokerage account) are exempt, at most 100,000 PLN together, savings first; other assets get no exemption. The rate is 0.85% for 2027, then 19% of the NBP reference rate in force on 31 October of the year before, at least 0.1%, rounded down to two decimal places; the exemptions rise with CPI from 2030. Income inside OKI is outside PIT, so no Belka tax is due.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-127#1
