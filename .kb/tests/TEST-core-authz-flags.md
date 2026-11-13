---
title: '--authz defaults to the policy, and SpiceDB''s key never shows'
status: active
sourceFile: cmd/core/authz_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.853Z'
id: TEST-core-authz-flags
type: test
---
The flags and environment of serve: the policy is the default and needs nothing of SpiceDB; spicedb needs an endpoint and the preshared key from the environment, refuses an unknown mode and a key given as a flag; the key shows in no format verb, JSON or log line of the configuration; the policy wires identityapp.Policy and no syncer; spicedb wires its authorizer and syncer, and with nothing listening refuses every check with an error. A false pass would be a redaction that hides only one formatting, which the table of verbs and the JSON log would catch.