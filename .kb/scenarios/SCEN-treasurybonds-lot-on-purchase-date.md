---
title: A bond lot is a deposit and a buy on its purchase date
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:34.043Z'
id: SCEN-treasurybonds-lot-on-purchase-date
type: scenario
---
Given lots entered on an account never seeded, in any order, when the treasury-bonds connector maps them, then each is a deposit of its cost and a buy on its own purchase date keyed by series and date, and a deleted lot deletes both; through the real core the lot is two new transactions and entering it again changes nothing; only an account seeded with opening lots keeps them.