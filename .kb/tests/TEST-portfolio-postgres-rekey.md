---
title: The model portfolio's target of a provisional instrument moves to its registry ISIN
status: active
sourceFile: internal/portfolio/postgres/rekey_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:20.724Z'
id: TEST-portfolio-postgres-rekey
type: test
---
Against a disposable PostgreSQL (task test:pg), a model targeting the provisional instrument has its target renamed in place, and one targeting both has the provisional target's weight joined to the registry's so the weights sum to 100; re-keying twice gives the same model each time.