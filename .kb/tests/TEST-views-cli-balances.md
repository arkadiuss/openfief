---
title: openfief check balances prints the report as a table or JSON and fails when a balance assertion fails
status: active
sourceFile: internal/views/cli/balances_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:09.204Z'
id: TEST-views-cli-balances
type: test
---
TestRunTable and TestRunJSON run the command against a fake core pod that answers with the JSON the core writes for a synthetic report: the core is asked for check balances, with --all only when given; the table is compared whole, the JSON read back strictly and compared with the report; a failed assertion is printed as FAIL and the command returns ErrBalancesFailed after printing. TestRunRefuses checks usage errors, a refused exec, and output of another version, with an unknown field or with wrong counts, printing nothing; TestWriteTableEscapes quotes a symbol with control and bidirectional characters.