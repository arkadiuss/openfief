---
title: Only rows OpenFief can't book count as unsupported
status: active
sourceFile: internal/ingest/problem_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:59.819Z'
id: TEST-ingest-problem
type: test
---
Pure table test of ingest.CountUnsupported: no problems, skipped rows only, and a mix, where only the unsupported rows count.