---
title: The owner's accounts carry the mark of an account opened on its seeding day
status: active
sourceFile: internal/ingest/app/opened_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T08:22:22.710Z'
id: TEST-ingest-app-opened
type: test
---
Runs ingest's Accounts use cases over in-memory fakes. A new account takes the mark, a change that doesn't name it keeps it, changes that name it clear and set it again, each stored with the account, and a read session's change or one of an account that is not stored is refused with nothing saved.