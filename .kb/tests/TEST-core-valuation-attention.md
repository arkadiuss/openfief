---
title: The core's adapter turns valuation's attention port into attention items
status: active
sourceFile: cmd/core/valuationattention_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:31.453Z'
id: TEST-core-valuation-attention
type: test
---
valuationAttention over a fake of the attention items: a raise of a market-data gap item becomes the condition with its kind, subject and detail, a clear names the kind and subject, and the store's failure reaches the caller wrapped, so a revaluation that can't raise its item fails.