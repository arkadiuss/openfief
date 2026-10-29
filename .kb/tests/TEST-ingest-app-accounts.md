---
title: The owner creates accounts with their connector, broker account, tax wrapper and fee rule; nobody else can
status: active
sourceFile: internal/ingest/app/accounts_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:19.537Z'
id: TEST-ingest-app-accounts
type: test
---
Use-case tests over in-memory fakes and the Polish pack: accounts get IDs from names, the pack's default wrapper taxable or the one given, such as an IKZE with a 0.29% at least 14 PLN fee rule, recorded through the terms port; a wrapper the pack lacks and a 100% fee rate are invalid accounts and store nothing; updates keep or change the terms; sleeve, instance and broker account checks; owner only.