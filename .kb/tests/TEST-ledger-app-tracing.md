---
title: Ledger imports run in a ledger.import span naming only how it imports
status: active
sourceFile: internal/ledger/app/tracing_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:51.638Z'
id: TEST-ledger-app-tracing
type: test
---
Under a root span, a successful import, an import into an unknown account, a seed and an addition of opening lots through app.Ledger over the in-memory store each produce one ledger.import child span whose only attribute is the mode (import, seed, opening_lots), with an error status, without a description, exactly when the call failed.