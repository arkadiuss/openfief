---
title: The ledger re-keys every account's holding of an instrument in one transaction, once
status: active
sourceFile: internal/ledger/app/rekey_test.go
origin:
  kind: agent
  recorded_at: '2026-11-07T22:44:18.113Z'
id: TEST-ledger-app-rekey
type: test
---
Imports a provisional series' buys and a coupon in two accounts plus another instrument into an in-memory store, re-keys it: no account holds the provisional instrument, acc-1's transactions are the re-keyed ones, each key has two versions, the import hook saw three new versions from each account's first day; re-keying again appends nothing and runs no hook; an invalid instrument is refused. Names.RekeyInstrument carries names once.