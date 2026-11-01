---
title: OKI's tax from accounts' days per class
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T18:33:16.781Z'
id: SCEN-pl-oki-tax-computation
type: scenario
---
Given OKI accounts' days with end-of-day values per class and payments, when the pack computes the year's tax, then each account's averages take the payments up to withdrawals in proportion to the classes, the exemptions apply savings first, and the rest is taxed at the year's rate, or no rate or tax while it isn't known.