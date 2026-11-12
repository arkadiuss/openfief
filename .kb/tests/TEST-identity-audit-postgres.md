---
title: The audit table refuses UPDATE, DELETE and TRUNCATE, takes no free text, and commits with its change
status: active
sourceFile: internal/identity/postgres/audit_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:22.433Z'
id: TEST-identity-audit-postgres
type: test
---
Under task test:pg on a disposable PostgreSQL. TestAuditLogIsAppendOnly runs UPDATE, DELETE and TRUNCATE as an ordinary role that owns the database, which fails with insufficient privilege, and as a superuser, which the trigger refuses, including statements that match no row. TestAuditLogRefusesFreeText inserts around the domain and checks the table's CHECKs. TestAtomicallyRollsBackStateAndEntry and TestAuditFailsClosed check that a change and its entry roll back together. TestAuditTrailOfOwnerActions runs the use cases over the real store and checks the entries and that no secret is in the log. TestAuditEntriesFilterAndPage checks filters and cursor paging.