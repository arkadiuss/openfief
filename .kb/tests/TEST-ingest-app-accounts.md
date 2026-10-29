---
title: The owner creates accounts with their connector and broker account; nobody else can
status: active
sourceFile: internal/ingest/app/accounts_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:19.537Z'
id: TEST-ingest-app-accounts
type: test
---
Use-case tests of ingest app Accounts over in-memory fakes with one stored account and two synthetic broker accounts: creating an account from a reported broker account takes its institution, binds the instance and maps it; a connector that names its own accounts is only bound; no connector binds nothing; a taken ID, no ID, no institution, an unknown sleeve, a malformed instance, a broker account without its instance, never reported or mapped elsewhere are refused before anything is stored; update and bind work and refuse unknown accounts; refused sessions store nothing.