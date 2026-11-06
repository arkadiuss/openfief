---
title: Every coupon since a lot's purchase is booked on every sync, once
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T09:18:20.537Z'
id: SCEN-treasurybonds-coupons-every-sync
type: scenario
---
Given an account never seeded holding a COI lot entered after two of its yearly coupons fell due, when the treasury-bonds connector syncs, then both coupons are booked as interest on their due dates, keyed lot:<series>:<purchase date>:coupon:<due date>, with Belka tax by the account's wrapper, and a second sync delivers the same keys with the same content, so nothing new is booked (D-143).