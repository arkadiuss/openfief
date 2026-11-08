---
title: Each held bond series' gross accrued value equals the reference valuer's and the hand calculation's to 0.01 PLN
status: active
tags:
  - acceptance:p10-5
origin:
  kind: agent
  recorded_at: '2026-11-08T00:35:36.784Z'
id: SCEN-e2e-bond-values-reference
type: scenario
---
Given the data path's taxable treasury bond account holding a TOS0329 lot (10 bonds bought 2026-03-10, 5.65%), an EDO0534 lot (20 bonds bought 2024-05-14, 6.50% then CPI plus 1.75%) and a COI1029 lot (30 bonds bought 2025-10-06, 5.60% then CPI plus 1.50%), with the demo dataset's CPI, when the core values the account, then each series' stored gross accrued value on 2026-03-31 and 2026-11-02 equals the value github.com/maciekmm/obligacje (MIT) at commit 8e1c495 computes offline on the same terms, and the hand calculation from the issue letters' formulas, to the grosz.