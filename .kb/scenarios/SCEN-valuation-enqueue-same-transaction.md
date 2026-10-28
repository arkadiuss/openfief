---
title: An input and its revaluations commit or roll back together
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:56:15.709Z'
id: SCEN-valuation-enqueue-same-transaction
type: scenario
---
Given an empty database, when a price is saved with a revaluation of acc-1, then the price, acc-1's pending date and a River job are all committed; when a batch of prices fails on its second row, then no price, pending date or job is stored; when a ledger import enqueues through its own transaction and then fails, then no job remains.