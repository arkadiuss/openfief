---
title: Under SpiceDB, core token issue with the household grant issues nothing when SpiceDB does not answer, and never leaves a token whose grant is not in SpiceDB
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T15:35:04.654Z'
id: SCEN-core-token-issue-atomic-under-spicedb
type: scenario
---
Given the core's own command run with --authz=spicedb and --grant-household, when SpiceDB does not answer within --spicedb-wait, then it issues nothing: no token, no grant, no audit entry, no write to SpiceDB, no secret on stdout, and an exit with an error that says so; when SpiceDB answers, then the token, its household grant and the entries of both are stored in one transaction, the grant is written to SpiceDB and the secret is printed first; and when the write of the grant still fails after the store, then the token is revoked and its grant ended, each with its own audit entry, no secret is printed and the error says the token was issued and revoked because its grant could not reach SpiceDB. A token issued without the grant needs nothing of SpiceDB and is issued whether it answers or not. Assumes a SpiceDB that fails whole calls, not one that takes a write and loses it.