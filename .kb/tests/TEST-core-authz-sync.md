---
title: The core passes placements on to SpiceDB and reconciles it at start
status: active
sourceFile: cmd/core/authzsync_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:54:34.833Z'
id: TEST-core-authz-sync
type: test
---
With a recording fake of SpiceDB's permissions service: a placement through cmd/core's observer writes the account's household and sleeve under escaped IDs, a removal writes nothing when no sleeve relationship is held, and the start's reconcile writes the household's owner and each account and sleeve the loader names, and does nothing with the policy. A false pass would be an observer that swapped account and sleeve IDs, which the exact write list would catch.