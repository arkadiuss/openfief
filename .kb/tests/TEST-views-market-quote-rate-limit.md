---
title: The market quote limit per minute
status: active
sourceFile: internal/views/app/quotes_test.go
verification_scope: unit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-views-market-quote-rate-limit
type: test
---
Over a fake source and a clock the test moves: two lookups pass, the third is refused without reaching the source, and the next minute counts afresh. A false pass would be a limiter that counts after asking the source.
