---
title: A statement upload and a bond lot entry and deletion ask the connector before their transaction
status: active
sourceFile: internal/ingest/app/nonetwork_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:54.886Z'
id: TEST-ingest-app-no-network-in-change
type: test
---
TestStatementUploadMakesNoNetworkCallInsideItsTransaction, TestBondLotEntryMakesNoNetworkCallInsideItsTransaction and TestBondLotDeletionMakesNoNetworkCallInsideItsTransaction run each change under an auditor that marks the context of its transaction, and check the connector and its manifest were called outside it and that the import and the record of the batch ran inside it.