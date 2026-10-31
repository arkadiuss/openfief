---
title: A sync of an instance bound to accounts carries what each holds, a market-data connector's none
status: active
sourceFile: internal/ingest/app/held_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:09.779Z'
id: TEST-ingest-app-held
type: test
---
Unit tests of ingest/app.HoldingsScope with fake ledger, seeding, bond and wrapper ports: each bound account once in ID order, its go-live only when seeded, its recorded tax wrapper or the pack's default, its holdings and open lots today, the terms of its bonds once; a failing ledger fails the scope. A scheduler test checks a bound instance's sync carries the held accounts (D-117) and an unbound one's carries none, through the real Scheduler with an in-memory store.