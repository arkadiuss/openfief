---
title: A revaluation runs in a valuation.revalue span with no attributes
status: active
sourceFile: internal/valuation/app/tracing_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T03:05:53.540Z'
id: TEST-valuation-app-tracing
type: test
---
Under a root span, a successful revaluation over the package's fake ledger and market store and one over a failing log each produce a valuation.revalue child span with no attributes, the second with an error status and no description.