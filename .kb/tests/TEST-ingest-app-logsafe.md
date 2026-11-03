---
title: Logged ingest failures carry no amounts, ISINs or account numbers
status: active
sourceFile: internal/ingest/app/logsafe_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T01:05:17.238Z'
id: TEST-ingest-app-logsafe
type: test
---
TestLogSafe describes synthetic failures that carry an amount, an ISIN and an IBKR-style account number, known ones by their sentinel's text and others by their innermost type, and checks no description matches a deny-list regex of those values. TestSchedulerLogsNoValues drives the sync scheduler, with in-memory fakes, through each failure and checks its text log holds one failure line each and nothing the regex matches, with the time left out since its fraction looks like an amount.