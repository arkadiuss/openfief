---
title: core check ledger writes the whole-log check as JSON and logs only counts
status: active
sourceFile: cmd/core/checkledger_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:27.420Z'
id: TEST-core-check-ledger
type: test
---
TestParseCheckLedger checks the command line, an optional --account and nothing else; TestCheckLedger runs core check ledger --account on a real PostgreSQL database holding a synthetic deposit and buy, the deposit then corrected, and checks the JSON report: balanced, three versions, one superseded and listed, sums of zero, and a log line with counts only that names no account, key or amount.