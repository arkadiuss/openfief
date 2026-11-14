---
title: The count of jobs that keep failing reads only the unfinished jobs of one kind
status: active
sourceFile: internal/valuation/postgres/changed_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:58.560Z'
id: TEST-valuation-postgres-jobs-failed
type: test
---
Under task test:pg: TestJobsFailedAtLeastCountsTheJobsThatKeepFailing checks the count of jobs of a kind that failed at least some attempts and are still waiting, leaving out jobs that failed fewer, jobs of another kind and finished ones.