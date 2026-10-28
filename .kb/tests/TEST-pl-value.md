---
title: Gross accrued and net-if-redeemed value of synthetic bond lots
status: active
sourceFile: internal/jurisdiction/pl/value_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:17.865Z'
id: TEST-pl-value
type: test
---
Table tests value synthetic lots of EDO, COI, ROR, OTS and TOS against hand-computed figures shown in each case's comment: accrual over days elapsed per actual days in the period, yearly compounding for capitalized types, the current coupon for coupon types, rounding to grosze per bond, the early-redemption fee capped at accrued interest (and uncapped for COI and ROR after the first period), no fee at maturity, Belka tax on interest minus fee, and no tax in IKE or IKZE. Error cases cover a date before purchase, a series mismatch, fractional and zero quantities, bad terms, an unknown wrapper and a missing CPI reading. Every amount is compared as PLN money with go-cmp.