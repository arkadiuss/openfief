---
title: Which connector a quote is asked of
status: active
sourceFile: internal/ingest/app/quotes_test.go
verification_scope: unit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-ingest-market-quote-source
type: test
---
Over fake instances, declarations, ranking and lookup: the order asked, the pass-over of a connector that takes no quotes, the stop at one that is down or finds nothing, the exclusion of an undeclared one and the failures of the lists. A false pass would be a fake that answers for every instance.
