---
title: 'On PostgreSQL: opinions off until turned on, on and off round trip, a failing hook stores nothing.'
status: active
sourceFile: internal/modelrouter/postgres/store_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: TEST-models-opinions-store
type: test
---
The opinions setting is stored by the core (D-225) in its own singleton table, and an owner change is audited in the transaction that stores it, so a store that keeps a change whose log entry failed would break the audit rule. This test runs the store against PostgreSQL because the table, its default and the rollback are the database's.

## Context

Asked for by the design session that handed over D-225 and D-229, 2026-11-21: the setting is off by default, stored core-side and audited.

## Source

> The toggle is owner-only, stored core-side, audited, OFF by default.
