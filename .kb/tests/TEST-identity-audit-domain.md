---
title: An audit entry is a record of IDs and closed values and refuses secrets, names, amounts and sentences, a grant entry naming its holder then its target
status: active
sourceFile: internal/identity/audit_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T19:17:22.172Z'
id: TEST-identity-audit-domain
type: test
---
Table tests of identity.AuditEntry, its actors, actions, outcomes and references, pure functions. The actions, outcomes and reference kinds are closed sets, including grant_created and grant_revoked and the references to a person, a view plugin, an account, a sleeve and the household. Each actor and reference must be an ID of a fixed shape: no name, email address, sentence, secret or amount is accepted in any position, and a message never quotes the rejected value. The filter and its page size are checked at their edges.