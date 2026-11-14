---
title: The revaluation client takes other contexts' jobs
status: active
sourceFile: internal/valuation/postgres/jobs_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:23.984Z'
id: TEST-valuation-postgres-jobs
type: test
---
Tests of NewClient with Config.Jobs: a job on a queue of its own is registered once, and one on the revaluations' queue, without a queue, workers or registration, or twice on one queue is refused, a failing registration passed on.