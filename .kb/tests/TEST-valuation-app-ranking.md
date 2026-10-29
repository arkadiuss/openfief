---
title: The owner ranks sources, and a changed ranking revalues every account
status: active
sourceFile: internal/valuation/app/ranking_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T18:00:41.656Z'
id: TEST-valuation-app-ranking
type: test
---
Use-case tests of app.Rankings over an in-memory store and two accounts: recorded sources rank EODHD first for prices, the owner's order replaces it and enqueues a revaluation of both accounts from their first transaction, the same order again enqueues nothing; another session, a source twice, no kind and a failing store are refused; a nameless source isn't recorded.