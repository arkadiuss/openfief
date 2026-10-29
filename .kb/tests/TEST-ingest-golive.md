---
title: A sync's batches follow the bound accounts' seeding, and a sync books nothing a seeding holds
status: active
sourceFile: internal/ingest/golive_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T16:57:01.267Z'
id: TEST-ingest-golive
type: test
---
Table tests of ingest.PlanSync over bound accounts and go-live dates: transactions alone when none is bound or all are seeded, opening alone when none is, opening then transactions when some are, and a seeded account of another instance not counting; and of ingest.KeepFromGoLive: in opening mode a seeded account's opening lot and opening cash become skipped problems while a new account's seeding stays, in transactions mode every transaction of a seeded account on or before its go-live date is left out, and the input mapping is unchanged. A false pass would need a mixed plan to drop its transactions batch.