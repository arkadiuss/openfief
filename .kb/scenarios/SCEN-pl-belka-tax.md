---
title: Belka tax is 19% of interest after the fee
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:29.511Z'
id: SCEN-pl-belka-tax
type: scenario
---
Given a taxable lot of 10 EDO bonds with 30.20 PLN accrued interest and 20.00 PLN of early-redemption fees, when it is valued as if redeemed today, then the Belka tax is 19% of 10.20 = 1.938, rounded half up to 1.94 PLN, and the net value is 1030.20 - 20.00 - 1.94 = 1008.26 PLN; given the same lot bought by exchange at 99.90 PLN per bond, then the 1.00 PLN discount is taxed too and the tax is 2.13 PLN; given interest below the fee, then there is no tax.