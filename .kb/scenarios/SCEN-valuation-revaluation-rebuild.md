---
title: A revaluation rebuilds the holding valuations from a date to today
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:36.769Z'
id: SCEN-valuation-revaluation-rebuild
type: scenario
---
Given an account that bought 10 ETF units on 2026-03-02 and stored prices and FX rates, when it is revalued from 2026-03-02 with today 2026-03-06, then it stores one holding valuation per day, keyed by account, instrument, date and basis; when it is revalued again from any date, then the stored rows are identical; when the date is before the first transaction, then it starts at the first transaction.