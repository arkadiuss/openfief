---
title: The performance query starts at go-live and ends at the latest valued-through date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:44.028Z'
id: SCEN-views-performance-query
type: scenario
---
Given synthetic accounts with transactions and valuation statuses, when the performance is asked for without dates, then it runs from the first transaction's date to the latest valued-through date of any account, an earlier from starts at go-live, an account never valued counts as valued through the day before its first transaction, and a pending or running revaluation is reported.