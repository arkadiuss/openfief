---
title: The syncer writes the schema, the household, account, sleeve and grant relationships, removes the ones the core no longer wants, and a second reconcile writes nothing
status: active
sourceFile: internal/identity/spicedb/syncer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T04:07:16.274Z'
id: TEST-identity-spicedb-syncer
type: test
---
Runs the Syncer over an in-memory SpiceDB. Placement, move and removal of an account write and remove its household and sleeve relationships in one atomic write. Reconcile writes the embedded schema first, the owner, the demo sessions' grant when they exist, and every account's household and sleeve, and removes those of a deleted account or a sleeve left; it also writes each grant in force as a relationship (a viewer of an account or a sleeve, a reader of the household, for a person, a script token or a view plugin) and removes a grant relationship the table no longer holds, an ended one or one written by hand, so SpiceDB is exactly what the core's tables say; a second Reconcile writes nothing. ProjectGrants and UnprojectGrants are idempotent, escape IDs SpiceDB cannot take, and refuse a kind of holder or target the schema lacks without writing. Every failure comes back so the caller can say the change is not in SpiceDB yet; a schema SpiceDB refuses stops the reconcile before any relationship is written.