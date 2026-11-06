---
title: An item names the account, instrument or instance it is about
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:38.760Z'
id: SCEN-attention-item-scope
type: scenario
---
Given attention items of every kind, when their scope is read from kind, subject and detail, then a balance assertion names its account and holding's instrument, a market-data gap its account and a prices series' instrument, a price check its instrument, and items tied to no account only their connector instance, never a broker account number; a setting and a broken transfer name nothing.