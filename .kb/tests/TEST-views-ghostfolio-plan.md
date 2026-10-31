---
title: The reconcile plan keeps, imports, reverts and deletes exactly what differs, and imports only after a clean dry run
status: active
sourceFile: internal/views/ghostfolio/plan_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:12.541Z'
id: TEST-views-ghostfolio-plan
type: test
verification_scope: unit
---
Pure table tests of MakePlan over two owned accounts: a first sync creates both and imports everything; an unchanged state plans nothing; a missing activity is imported; an edited quantity is reverted by importing the desired one and deleting the edited one; a stale activity and a second copy are deleted; an owned account that should hold nothing is cleared with a filtered bulk delete; an activity moved between owned accounts is deleted before the import it would block; an activity outside the owned accounts is untouched. TestCheckPreview refuses a dry run with a duplicate, another symbol or account, a missing or a replaced activity; TestBatches checks batches carry the assets their activities use.