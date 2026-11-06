---
title: A broken transfer's item is tied to both legs' accounts
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T03:42:55.521Z'
id: SCEN-views-transfer-broken-accounts
type: scenario
---
Given a confirmed transfer whose withdrawal is in acc-1 and whose deposit is in acc-2 and an open transfer-broken item about it, when the read API lists attention items, then the item names acc-1 and acc-2, the withdrawal's first, an account filter for either finds it and one for another account doesn't, and a caller who may not view acc-1 or may not view acc-2 doesn't see it.