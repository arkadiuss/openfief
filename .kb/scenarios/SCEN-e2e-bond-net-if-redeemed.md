---
title: Each held bond series' net and liquidation values and the account's totals equal the hand calculation
status: active
tags:
  - acceptance:p10-5
origin:
  kind: agent
  recorded_at: '2026-11-08T00:35:38.019Z'
id: SCEN-e2e-bond-net-if-redeemed
type: scenario
---
Given the same bond account, when the core values it, then each series' net value (gross less 19% Belka tax on its interest) and liquidation value (gross less the early-redemption fee, capped at the accrued interest except for COI after its first period, less Belka tax on the interest after the fee) on both days, and the account's gross, net and liquidation totals with the cash of COI's first coupon after tax, equal the hand calculation to the grosz.