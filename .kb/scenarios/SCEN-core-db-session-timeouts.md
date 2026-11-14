---
title: A pool's connections start with the timeouts and the database ends an idle transaction
status: active
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:04.615Z'
id: SCEN-core-db-session-timeouts
type: scenario
---
Given the core's database pool, when a connection starts, then its statement_timeout and idle_in_transaction_session_timeout are the defaults, the environment's value or the connection string's, and a transaction left idle past the idle timeout is ended by the database and rolled back.

## Context

The timeouts are runtime parameters of each connection, so they hold for the bundled database, the demo's and the owner's own, with no role or server setting. A connector's delivery sets its own statement timeout for its transaction. The demo's loader connects with its own connection string and sets neither.

## Source

> DB ROLE TIMEOUTS: set idle_in_transaction_session_timeout and a sane statement_timeout for the core's DB connections, configurable with defaults, in the place that also applies to the demo and the bundled DB.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18