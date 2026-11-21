---
title: Which callers may read a market quote
status: active
sourceFile: internal/views/connectapi/quotes_test.go
verification_scope: integration
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md), and the design session's ruling on it
  recorded_at: '2026-11-21T14:00:00.000Z'
id: TEST-views-market-quote-callers
type: test
---
Through the real read API handler over a fake quote source: the owner and a script token are answered, a view plugin and a demo session are permission denied without the source being asked, and at the views and the identity policy each holder's answer, a viewer's included, is the one D-226 gives. A false pass would be a handler that checks the holder after asking the source.
