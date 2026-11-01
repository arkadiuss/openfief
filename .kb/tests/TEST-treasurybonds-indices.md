---
title: The index sync delivers each index after its own watermark, CPI a few years a sync
status: active
sourceFile: internal/ingest/treasurybonds/indices_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:17.396Z'
id: TEST-treasurybonds-indices
type: test
verification_scope: unit
---
Table tests with in-memory fakes of GUS and NBP: the watermark round-trips and rejects other connectors' watermarks; ReferenceReadings gives the rate in force on each month's reference-rate day, waits while that day is today and skips months before the first change; IndexSync.Readings delivers CPI and reference rates from the start or after each index's watermark, asks GUS for at most four years and NBP only when a month is due, and fails on a bad watermark or an outage. Synthetic rates and CPI values only.