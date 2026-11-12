---
title: The syncer writes the household, account and sleeve relationships and leaves grants alone
status: active
sourceFile: internal/identity/spicedb/syncer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.764Z'
id: TEST-identity-spicedb-syncer
type: test
---
Over an in-memory SpiceDB: a placement writes the account's household and sleeve, a move is one atomic write that removes the old sleeve, removing an account from its sleeve keeps its household, reconcile writes the missing relationships and removes those of deleted accounts and left sleeves, follows the demo-session flag, writes nothing the second time, batches over 500 updates, and in every case leaves the owner's grants (household readers, account and sleeve viewers). Every failure, timeout and empty ID comes back as an error. A false pass would be a fake that ignored the filters of the reads, which the grants assertions would catch.