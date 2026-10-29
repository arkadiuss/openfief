---
title: The Polish pack values a holding, a bond and cash in each tax wrapper
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:45.265Z'
id: SCEN-pl-values-per-wrapper
type: scenario
---
Given a holding worth 10 000 PLN on a cost of 8 000 with a sale cost of 29, a bond lot of 1 030.20 PLN with 30.20 of interest, and 500 PLN of cash, when the pack values them in taxable, IKE, IKZE and OKI, then net is 9 620, 10 000, 9 000 and 10 000 for the holding and liquidation 9 596.51, 9 596.51, unknown without the owner's marginal rate (6 780.28 at 32%) and 9 971; IKZE cash nets 450; no value ever exceeds gross.