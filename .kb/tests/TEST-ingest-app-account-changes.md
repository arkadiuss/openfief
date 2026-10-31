---
title: The owner deletes an empty account, remaps a broker account its account holds nothing from, and alone lists broker accounts
status: active
sourceFile: internal/ingest/app/accountchanges_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:02.229Z'
id: TEST-ingest-app-account-changes
type: test
---
Unit tests of ingest/app.Accounts with fake ledger, ingest and valuation ports: deleting an account with transactions is refused and forgets nothing; an empty one is deleted by the ledger first, then forgotten by ingest and valuation; a read session may not delete. Table cases of Remap: an unmapped broker account maps and binds; one whose account holds transactions from its instance is refused; another instance's transactions don't count; an unknown broker account or target account is refused. BrokerAccounts lists all or unmapped for the owner and refuses a read session.