---
title: A new ranking revalues every account once
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:52.086Z'
id: SCEN-valuation-ranking-revalues
type: scenario
---
Given two accounts, when the owner ranks ibkr-main above prices-eu for prices, both accounts are revalued from their first transaction in the transaction that stores the order, and reads return ibkr-main's price; saving the same order again enqueues nothing.