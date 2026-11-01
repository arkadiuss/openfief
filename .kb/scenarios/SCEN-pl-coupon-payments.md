---
title: A coupon series pays each period's coupon at the period's end
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:17.108Z'
id: SCEN-pl-coupon-payments
type: scenario
---
Given a COI or ROR lot's series terms and its period index readings, when the Polish pack dates and computes its coupons, then each falls due on the first day of the next interest period, except the last period's, and one bond is paid the nominal times the period's rate times months over twelve, rounded half up to grosze.