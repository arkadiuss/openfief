---
title: An account never seeded is paid from purchase in opening batches
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:38.588Z'
id: SCEN-treasurybonds-payouts-from-purchase
type: scenario
---
Given an account never seeded holding a coupon lot entered on its purchase date, when the connector syncs in opening mode, then it books every coupon since purchase and pays out those after its payout setting's day; in transactions mode it books only accounts seeded with opening lots, from their go-live.