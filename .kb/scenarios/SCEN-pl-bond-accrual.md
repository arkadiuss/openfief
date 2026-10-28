---
title: Bond interest accrues per actual day and rounds to grosze per bond
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:22.415Z'
id: SCEN-pl-bond-accrual
type: scenario
---
Given an EDO bond with a 6% first-year rate bought on 2024-03-15, when it is valued on 2024-09-15 (184 of 365 days), then it is worth 100 x (1 + 0.06 x 184/365) = 103.02 PLN; given a COI bond at 6.5% bought on 2025-01-10, when valued on 2025-07-10, then it is worth 100 PLN plus the current coupon 3.22 PLN. Assumes synthetic rates.