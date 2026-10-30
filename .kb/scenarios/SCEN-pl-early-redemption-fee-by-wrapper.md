---
title: A bond lot's early-redemption fee follows its wrapper and its cap
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:47.406Z'
id: SCEN-pl-early-redemption-fee-by-wrapper
type: scenario
---
Given an EDO lot of 10 bonds with a 2.00 PLN fee valued mid first year, when it is held in taxable, IKE, IKZE or OKI, then its fee is 20.00 in taxable and OKI and nothing in IKE and IKZE; a fee above the accrued interest is capped at it except for coupon types after their first period; no fee at maturity.