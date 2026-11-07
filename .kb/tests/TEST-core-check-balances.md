---
title: core check balances writes the report from a real database and logs only counts
status: active
sourceFile: cmd/core/checkbalances_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:09.829Z'
id: TEST-core-check-balances
type: test
---
TestCheckBalances starts the core on a fresh PostgreSQL under task test:pg, imports a synthetic deposit, a buy of four units and the broker's synthetic report of the cash and five units, and records an XTB upload batch that left one open CFD row out. Running core check balances writes JSON that ReadJSON accepts: the cash passes, the holding fails a unit off, and the instance's unsupported row is listed. The JSON log says the check ran with its counts and names no account, instrument or amount.