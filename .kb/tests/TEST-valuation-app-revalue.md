---
title: Revalue rebuilds an account's holding valuations from its log and stored inputs
status: active
sourceFile: internal/valuation/app/revalue_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:28:23.924Z'
id: TEST-valuation-app-revalue
type: test
---
Uses an in-memory ledger log (acc-1 buys 10 synthetic ETF units on 2026-03-02) and an in-memory store. Revaluing through 2026-03-06 stores five rows and is valued through today with every input, through 2026-03-01 when the first day's price, all FX rates or the first two days' rates are missing; stored rows carry the close times quantity in USD and in PLN; running again from any date leaves identical rows; the start moves to the first transaction; an earlier gap is kept until its own input arrives. A synthetic EDO holding is valued on the configured basis and tax wrapper exactly as pl.ValueLot gives; the market basis, a failing log and a failing store are errors.