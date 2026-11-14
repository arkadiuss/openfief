---
title: A viewer's Ask box question is entered as the viewer's, not the owner's
status: active
sourceFile: cmd/core/askmodel_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:44.638Z'
id: TEST-core-ask-actor
type: test
---
TestAskSessionsAdmitAnyValidSession checks the Ask box's session port answers the owner for the owner's sessions and a viewer for a viewer's, by session ID; TestModelAuditRecordsCallsAsAuditEntries checks a viewer's call is entered as the viewer's, sent and ended; the models app's TestAskGenerateIsTheAskRolesCallForTheSession checks the router is called for the viewer.