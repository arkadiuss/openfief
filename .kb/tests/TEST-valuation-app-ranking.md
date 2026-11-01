---
title: The owner ranks sources, and a changed ranking revalues every account
status: active
sourceFile: internal/valuation/app/ranking_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:41.656Z'
id: TEST-valuation-app-ranking
type: test
verification_scope: unit
---
Use-case tests of Rankings over in-memory fakes: sources recorded from deliveries rank EODHD first by default, the owner's new order is stored and returned with each kind's gap threshold, defaulting to 7 and 92 days, and a change of ranking revalues nothing because every series keeps its recorded source (D-94); another session, a source ranked twice, an unknown kind and a failing store are refused, and a nameless source is never recorded.