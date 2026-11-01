---
title: The owner rejects a confirmed transfer a correction broke
status: active
sourceFile: internal/ledger/app/transferbroken_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:56.885Z'
id: TEST-ledger-app-transfer-broken
type: test
---
Over the ledger app with its in-memory store: confirms a pair, imports a corrected deposit, then confirming it again is refused as an invalid transfer while rejecting it succeeds as the pair's second version without checking the legs; rejecting a never-confirmed pair still checks them.