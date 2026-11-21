---
title: A held instrument's quote is its valuation source's latest price
status: active
sourceFile: internal/ingest/app/quotes_test.go
verification_scope: unit
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
id: TEST-ingest-market-quote-held-source
type: test
---
Over fake instances, ranking and a lookup that answers each source's own close: a held instrument is asked of its valuation source though another ranks first and its quote's latest close is that source's, an instrument not held is the ranked first source's, a held one whose source finds nothing is not asked of another, and one whose source is not running has no source. The valuation side, which source a ticker or ISIN names in the price series, is the pure function's table test. A false pass would be a lookup that always asks the ranked first.
