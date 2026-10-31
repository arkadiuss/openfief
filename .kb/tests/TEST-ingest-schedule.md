---
title: The sync policy validates and computes the next sync with jitter and a doubling backoff
status: active
sourceFile: internal/ingest/schedule_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:49:54.559Z'
id: TEST-ingest-schedule
type: test
verification_scope: unit
---
Table tests of the pure ingest package: SyncPolicy.Validate refuses a zero interval, a negative jitter, a zero first retry and a maximum retry below the first, and accepts no jitter; NextSync after a success adds the interval and the jitter clamped to zero and the policy's jitter, and after one, two, four, five and a thousand failures returns the first retry delay doubled per failure up to the maximum, without jitter. A false pass would be a backoff that never reaches the maximum, which the five-failure case rules out.