---
title: An audit entry is a record of IDs and closed values and refuses secrets, names, amounts and sentences
status: active
sourceFile: internal/identity/audit_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:22.172Z'
id: TEST-identity-audit-domain
type: test
---
Table tests of identity.NewAuditEntry, Ref, Actor and AuditFilter, in-process with no store. TestNewAuditEntry covers each valid action and each rejection: unknown action, outcome or actor kind, an owner named by a name or an email address, a reference kind the action does not name, a reference twice. TestAuditEntryRejectsSecretsAndFreeText feeds a script token secret, a session token, an enrollment code, an amount, a question's text, a long text and a newline into every reference kind and the owner actor and checks none gets in and no error quotes it. A validator that accepted a free-text field would fail it. The filter tests check the range, page size and cursor.