---
title: A confirmed pair's legs name the accounts they are in, the withdrawal's first
status: active
sourceFile: internal/views/pairs_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-06T03:42:54.258Z'
id: TEST-views-pairs
type: test
---
Table test of views.PairAccounts over synthetic transactions: legs in two accounts give both with the withdrawal's first, a deleted deposit gives only the withdrawal's account, both legs in one account give it once, and a pair with both legs deleted is left out. A false pass would key pairs other than the way a transfer-broken item's subject is written; the test looks pairs up by TransferPair.String.