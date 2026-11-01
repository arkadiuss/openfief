---
title: Coupons due after go-live are booked as interest once
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:19.049Z'
id: SCEN-treasurybonds-coupons-booked
type: scenario
---
Given a sync in transactions mode whose request says a seeded bond account holds an open COI or ROR lot with its series terms, when the treasury-bonds connector syncs, then each coupon that fell due after go-live and on or before today is an interest transaction keyed per lot and coupon date with Belka tax by the account's wrapper, a coupon whose reading isn't published waits as a problem, and syncing again changes nothing.