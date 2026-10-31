---
title: The ledger records transfer decisions in one transaction, for the owner only
status: active
sourceFile: internal/ledger/app/transfers_test.go
origin:
  kind: agent
  recorded_at: '2026-10-30T06:05:30.975Z'
id: TEST-ledger-app-transfers
type: test
verification_scope: unit
---
Imports a withdrawal and deposits into an in-memory store with the Postgres store's rules, then drives Ledger.DecideTransfer step by step: a confirmation stores version 1, repeating it appends nothing, a rejection supersedes as version 2 with version 1 kept, a second confirmation of a taken deposit is refused, legs too far apart, of the wrong kind or unknown are refused and nothing is stored. Transfers.Decide refuses a session that is not the owner's before storing anything, and two confirmations racing for one deposit leave one link.