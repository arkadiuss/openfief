---
title: Holdings are summed from current versions up to a date
status: active
sourceFile: internal/ledger/holding_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:38:51.913Z'
id: TEST-ledger-holding
type: test
---
A synthetic log holds an opening lot, a buy corrected from 10 to 12 units (two versions), a sale, a dividend, a buy and full sale of a second instrument, and a buy in another account. HoldingOn is table-tested per instrument and date: 0 before anything, 3 on go-live, 15 after the corrected buy (so version 1 is not counted and the other account is ignored), 11 after the sale and after the dividend (no units move), and 0 for the same ISIN on another venue symbol. HoldingsOn returns non-zero holdings in ISIN order and leaves out a sold-out one. A log with a repeated version fails with ErrInvalidLog from HoldingOn, HoldingsOn and LotsOn.