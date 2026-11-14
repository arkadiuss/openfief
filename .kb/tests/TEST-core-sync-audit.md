---
title: A sync's entry is the system's, only when it changed state, and the bind is the administrator's
status: active
sourceFile: cmd/core/outbox_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:43.091Z'
id: TEST-core-sync-audit
type: test
---
Under task test:pg: TestSyncEntryOnlyWhenTheSyncChangedState enqueues, over a real database, the entry of a sync that changed nothing (none), transactions, prices, bond terms, every kind, more accounts than an entry names and a sync that fails (none), and checks the actor, the instance, the batch, the counts and the time; TestConnectorBindAndSync runs core connector bind and core connector sync as the CLI does and checks the audit log holds the administrator's account_bound and the system's one changing sync, and nothing for the sync that brought nothing new or failed.