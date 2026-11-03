---
title: jobError hands River nothing, a snooze as it is, or the job and the failure's kind
status: active
verification_scope: unit
sourceFile: internal/valuation/postgres/joberror_internal_test.go
origin:
  kind: agent
  recorded_at: '2026-11-03T15:15:43.793Z'
id: TEST-valuation-postgres-job-error
type: test
---
Calls jobError with no failure, with River's snooze wrapped by the worker, with a valued failure wrapping a known error and with one wrapping an unknown error, and compares the results: nil, the snooze itself, the job with the known error's text, and the job with the innermost type. Any value in the text fails it.