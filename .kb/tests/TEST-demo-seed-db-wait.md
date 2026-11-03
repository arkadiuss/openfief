---
title: The demo loader waits for its database with a bounded, capped backoff
status: active
sourceFile: cmd/demo-seed/dbwait_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T17:14:50.597Z'
id: TEST-demo-seed-db-wait
type: test
---
Table tests drive waitForDatabase with a fake database that refuses a number of pings and a fake sleep that records each wait, so no real time passes. They check that a ready database is used at once, that refusals are retried after 0.5 s, 1 s, 2 s and so on up to a 5 s cap, that the waits never add up to more than the limit and a database still refusing then fails with errDatabaseNotReady wrapping the last refusal, and that a canceled wait stops at once.