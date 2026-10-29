---
title: The cross-check use cases validate, store and compare at the configured threshold
status: active
sourceFile: internal/valuation/app/crosscheck_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T05:37:05.650Z'
id: TEST-valuation-app-crosscheck
type: test
---
Table and scenario tests of app.CrossChecks over an in-memory store: NewCrossChecks takes 0.01 for an unset threshold, keeps 0.005, and refuses 1 and a negative threshold; RecordCrossCheckPrices stores valid synthetic IBKR marks, stores nothing of a batch with one invalid price and wraps ErrInvalidInput, and passes a store failure on; PriceChecks reads prices and cross-check prices from the store and, at a 0.3% threshold, matches an equal mark and flags a 0.4% one.