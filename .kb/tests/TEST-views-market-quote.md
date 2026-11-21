---
title: GetMarketQuote over Connect and REST, its errors, the settings flag and the period
status: active
sourceFile: internal/views/connectapi/quotes_test.go
verification_scope: integration
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: TEST-views-market-quote
type: test
---
Through the real read API handler over a fake quote source: the Connect and REST answers for the owner and a script token, the candidates, each error's status with a text that quotes nothing, GetSettings' market_quotes, and the views' period defaults and invalid requests, which never reach the source. A false pass would be a source that answers whatever it is asked.
