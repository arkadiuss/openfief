---
title: The core binds connector instances, syncs them through a real Connect server and schedules them on serve
status: active
sourceFile: cmd/core/connectors_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:34.037Z'
id: TEST-core-connectors
type: test
verification_scope: integration
---
Runs under task test:pg on a fresh database with the demo registry, a Connect server over h2c on 127.0.0.1 that answers Sync with a deposit only for the core's bearer token, and a fake cluster whose Plugin is Ready and whose Service points at that server. connector bind refuses an unknown account and binds the IBKR demo account; connector sync in transactions mode imports one deposit keyed plugin:ibkr-main:cash, a second sync finds it unchanged and stores the watermark it returned, an unbound instance is refused with ErrAccountNotBound and a sync outside a cluster fails. start with connectors on records the first, refused attempt in the ingest schema and shutdown waits for the scheduler. monthlyReading maps a reference rate to its month and refuses an unknown index.