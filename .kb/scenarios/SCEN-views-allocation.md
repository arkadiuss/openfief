---
title: The allocation splits the portfolio's PLN value by a chosen dimension
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:04.591Z'
id: SCEN-views-allocation
type: scenario
---
Given accounts at two institutions in two sleeves holding a USD ETF, a PLN stock, a treasury bond series and cash in USD and EUR, each valued in PLN on a date, when the allocation is asked for by asset class, currency, institution or sleeve, then every item lands in exactly one group named by its asset class (cash, treasury_bond, listed), its original currency, its institution or its sleeve, each group is the exact sum of its PLN values, the groups add up to the total, and the largest group comes first. A holding with no valuation on the date is reported as unvalued instead of counted.