---
title: The maintainer's River job and its weekly schedule
status: active
sourceFile: internal/portfolio/jobs/maintainer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:19.283Z'
id: TEST-portfolio-jobs-maintainer
type: test
---
Tests of WeeklyAt.Next across weeks, months and a clock change, the worker calling the run or the weekly turn, snoozing until the maintainer is built, and handing River an error with no text.