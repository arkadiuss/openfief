---
title: Read queries derive current transactions, holdings, lots and cash from the stored log
status: active
sourceFile: internal/ledger/app/query_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:16.577Z'
id: TEST-ledger-app-query
type: test
---
A synthetic account gets a deposit, two buys, a sale of 12 units, an FX conversion, and a deposit first booked to a second account and then corrected into the first. CurrentTransactions returns the current versions in date then key order, including the moved deposit, and none for the second account. On three dates, before any row, after both buys and after the sale, HoldingsOn, LotsOn and CashOn are compared in full: 15 then 3 units, the FIFO lot left from the second buy with 180 of its 300 USD cost, and cash per currency in code order (90 EUR, 2050 USD), with nothing before go-live. Every query for an unknown account fails with ErrUnknownAccount.