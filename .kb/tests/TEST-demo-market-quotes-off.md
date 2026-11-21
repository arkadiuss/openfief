---
title: The demo gets no market quote source
status: active
sourceFile: cmd/core/quotes_test.go
verification_scope: unit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-demo-market-quotes-off
type: test
---
With the demo flag and a discovery that finds an approved market-data connector, over views with no ports: the availability is false and a lookup is refused as unavailable before any connector or the ranking's store is touched. A false pass would be a source that is merely down.
