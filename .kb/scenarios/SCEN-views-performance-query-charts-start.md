---
title: The performance query starts on the charts' start, by default the latest seeding day
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:26.188Z'
id: SCEN-views-performance-query-charts-start
type: scenario
---
Given synthetic accounts with transactions and valuation statuses, one of them seeded from opening cash a day after the other's first deposit, when the performance is asked for without dates, then it runs from the seeding day, 100 on it, to the latest valued-through date of any account and names that day as the default start; a chosen earlier day is kept with the default still named, a day before the first transaction starts on it, an account never valued counts as valued through the day before its first transaction, and a pending or running revaluation is reported.