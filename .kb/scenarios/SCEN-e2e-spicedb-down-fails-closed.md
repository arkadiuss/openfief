---
title: With the core on SpiceDB, a stopped SpiceDB makes the read API refuse, and a started one makes it answer again
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.646Z'
id: SCEN-e2e-spicedb-down-fails-closed
type: scenario
---
Given the core started with --authz=spicedb and reconciled with a SpiceDB that holds no grant for a script token, when the token lists accounts, then none are listed, where the policy lists every account; when SpiceDB is stopped, then the read API refuses with 503 and the text authorization is unavailable and never answers 200, while the core's /healthz and /readyz still answer 200; when SpiceDB is started again, then the read API answers from the relationships PostgreSQL kept, and with the core back on the policy the token lists every account again. Assumes a script token as the caller, since the suite has no grants yet to give one an account (D-209).