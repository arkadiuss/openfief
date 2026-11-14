---
title: Every owner change writes exactly one entry, through the real core, and none without the log
status: active
sourceFile: cmd/core/ownerchanges_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:39.024Z'
id: TEST-core-owner-changes-audited
type: test
---
Under task test:pg, drives each owner procedure of the valuation, portfolio, ingest, ledger, attention and models contexts through a real core with an owner signed in, and checks each leaves exactly one ok entry of the owner's session and of the action; takes the audit log away and checks no change succeeds and the idempotent ones answer unavailable. TestEveryOwnerProcedureIsClassifiedForTheAuditLog lists the procedures of the contracts and fails for one in no list; TestCoveredElsewhereNamesTestsThatExist checks the tests it relies on exist. The upload, bond lot and ticker procedures are pinned by TestStatementService, TestBondLotService and TestIngestOwnerChangesAreAudited.