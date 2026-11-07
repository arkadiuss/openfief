---
title: Cash passes within 0.01 of its currency and holdings only with equal units
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:10.476Z'
id: SCEN-views-balance-check-tolerance
type: scenario
---
Given balance assertions an institution reported and the balances the ledger derived on their dates, when the owner's balance check runs, then each assertion shows reported, derived and reported minus derived; a cash assertion passes when the difference is at most 0.01 of its currency either way and fails past it; a holding assertion passes only when the units are equal; by default only the latest assertion of each account and ledger account is listed.