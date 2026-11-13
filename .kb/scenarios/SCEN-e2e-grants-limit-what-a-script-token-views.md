---
title: With the core on SpiceDB, a script token granted one sleeve views that sleeve's accounts alone, and a revoked grant views nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:20.856Z'
id: SCEN-e2e-grants-limit-what-a-script-token-views
type: scenario
---
Given the stack whose core answers from SpiceDB and an owner session, when a script token with no grant calls the read API, then it lists nothing and each account answers 404; when the owner grants it the defensive sleeve it lists that sleeve's account alone and its allocation by sleeve holds that sleeve alone, an account created in the sleeve afterwards is listed too and every account outside it answers 404; when the owner revokes the grant the next call lists nothing; and the UI plugin, installed before the core started on SpiceDB, still lists every demo account. The audit log shows each grant and revocation naming the token and the target by ID. Assumes D-209 and D-211.