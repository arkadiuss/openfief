---
title: The core is ready only after SpiceDB answered with its schema and stays ready through an outage, which it logs and counts once each way
status: active
sourceFile: cmd/core/authzmonitor_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T15:35:06.820Z'
id: TEST-core-authz-outage
type: test
---
With a scripted authorizer and a SpiceDB fake that can fail its reads: the monitor logs the change to unavailable once however many checks fail, the way back once, never for a refusal's error or the first answer; after an outage it asks SpiceDB on its own until it answers, logs the way back and stops asking, and does so again for a later outage; the gauge is absent before the first check, 1 while it answers and 0 during an outage, and the counter counts each change by the state it went to; the wrapper lists accounts exactly when the authorizer it wraps does; /readyz is 503 before the start's reconcile and after one SpiceDB refused, naming what it waits for, 200 after the one that succeeded, and 200 during an outage and after it; and a core without the key refuses as unavailable, its ping included. A false pass would be a readiness that follows the monitor, which the 200 during the outage would catch.