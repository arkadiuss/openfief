---
title: The time-weighted return chains daily returns across cash flows, transfers and gaps
status: active
sourceFile: internal/views/performance_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:27.083Z'
id: TEST-views-performance
type: test
---
Table tests over synthetic deposits, withdrawals, buys, opening lots and stored valuations of one ETF, with hand-computed indexes: a price move is a return; a deposit invested at the day's price, a deposit kept in cash, a withdrawal and an account seeded later are not; a USD transfer between two accounts is in transit at the day's rate in either booking order, and one further apart than the window is two cash flows; a missing, absent or stale valuation, the valued-through date and a currency without a rate stop the series with the named gap. TestMatchTransfers and TestExternalFlows check the matching rule case by case. A false pass would need the hand-computed values to share a mistake with the code.