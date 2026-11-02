---
title: CPI years served in full are kept, so later syncs get further
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T09:18:22.657Z'
id: SCEN-treasurybonds-cpi-years-kept
type: scenario
---
Given five COI lots bought a year apart, whose coupons and maturities need more years of CPI than one sync asks GUS for, when the connector syncs three times with the same CouponSync, then the first sync asks for four years and leaves the rest waiting as problems, the second asks only for the year it lacks and books every coupon and maturity, and the third asks GUS for nothing.