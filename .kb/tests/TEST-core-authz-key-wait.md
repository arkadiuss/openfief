---
title: A core without SpiceDB's key serves, refuses every check, reports not ready, and goes on when the key is there
status: active
sourceFile: cmd/core/authzkey_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T13:03:10.510Z'
id: TEST-core-authz-key-wait
type: test
---
The key gate over a key file in a temporary directory: with no file, an empty file and a file of whitespace, every check (MayAct, ActableAccounts) and every change the syncer would pass on (ProjectGrants, AccountPlaced) is refused with the pending error and the gate is not connected; with the key in the file the gate connects, and the polling run connects when the key appears later; /readyz answers 503 waiting for the key while /healthz answers 200, and the authorizer reports ready once the key is read; the wait is logged once and no line of the log shows the key; readKeyFile trims and treats a missing, empty or whitespace file as no key; serve gets the gate only with SpiceDB and no key in the environment; the commands that can't wait ask for the key at once. A false pass would be a gate that answers from the policy while it waits, which the refusal of the owner's own check would catch.