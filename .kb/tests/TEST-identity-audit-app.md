---
title: Each identity event writes its audit entry, fails closed without it, and only an owner session reads the log
status: active
sourceFile: internal/identity/app/audit_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:22.383Z'
id: TEST-identity-audit-app
type: test
---
Runs the auth use cases over the in-memory store, whose Atomically restores every table when its function fails, like a rolled back transaction. TestAuditTrail checks the exact entries of an enrollment, login, token creation and revocation, link, logout and sign out everywhere. TestAuditFailsClosed makes the audit write fail and checks each change returns the error and leaves no session, passkey or entry behind. TestRefusedSignInIsRecorded and TestRefusedSignInStaysRefusedWhenTheLogIsDown check the standalone entry for a refusal. TestAuditEntriesOwnerOnly checks a demo session, no session and a script token secret are refused. It proves the logic with a fake, not the database's rollback.