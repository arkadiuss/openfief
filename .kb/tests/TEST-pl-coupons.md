---
title: The Polish pack dates and computes each coupon of ROR, DOR and COI
status: active
sourceFile: internal/jurisdiction/pl/coupon_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:06.144Z'
id: TEST-pl-coupons
type: test
---
Pure table tests of pl.CouponDates and pl.Terms.Coupon on synthetic series terms: COI pays three yearly coupons and ROR eleven monthly ones, the last period's interest coming with the redemption; EDO and OTS pay none; a purchase in another month than the series names is refused. Amounts are worked by hand in each case's comment: N x r for COI, N x r / 12 for ROR rounded half up to grosze, a negative CPI floored at zero, a later period without its reading a *MissingIndexError. A false pass would need the hand-worked amounts themselves to be wrong.