---
title: With the core on SpiceDB, a stopped SpiceDB makes the read API refuse, and a started one makes it answer again
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.646Z'
id: SCEN-e2e-spicedb-down-fails-closed
type: scenario
---
Given the core started with --authz=spicedb, the default, and reconciled with a SpiceDB that holds no grant for a script token and the household grant for another, when the tokens list accounts, then the first lists none and the second every account; when SpiceDB is stopped, then the read API refuses with 503 and the text authorization is unavailable and never answers 200, while the core's /healthz and /readyz still answer 200; when SpiceDB is started again, then the read API answers again from the relationships PostgreSQL kept. Assumes script tokens as the callers, one issued with core token issue --grant-household (D-209).