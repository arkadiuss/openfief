---
title: The pool's connections start with the timeouts and an idle transaction is ended by the database
status: active
sourceFile: cmd/core/dbtimeouts_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:56.312Z'
id: TEST-core-db-timeouts
type: test
---
TestDBTimeoutsFromEnv and TestPoolConfigSetsTheTimeoutsOnEveryConnection check the defaults, the environment, zero and the refusal of a bad value, and that the connection string's own parameter wins; under task test:pg, TestPoolStartsEveryConnectionWithTheTimeouts reads both settings from a real connection and TestIdleTransactionIsEndedByTheDatabase checks a transaction left idle is ended with 25P03 and rolled back.