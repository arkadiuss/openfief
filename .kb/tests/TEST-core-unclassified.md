---
title: The core's adapter turns valuation's attention port into attention items
status: active
sourceFile: cmd/core/unclassified_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T19:57:18.771Z'
id: TEST-core-unclassified
type: test
---
TestValuationAttention drives valuationAttention, the adapter from valuation's AttentionRaiser port to the attention items, over a fake that records each call: raising instrument-unclassified for a synthetic ISIN must reach the items as the condition with the same kind, subject and detail, clearing it must name the same kind and subject, and a failing attention store must reach the caller as a wrapped error, so a revaluation that can't raise its item fails rather than dropping it (D-132, D-130).