---
title: '--authz defaults to SpiceDB, the policy is the opt-out that needs nothing of it, and SpiceDB''s key never shows'
status: active
sourceFile: cmd/core/authz_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-12T22:51:24.853Z'
id: TEST-core-authz-flags
type: test
---
The flags and environment of serve and the token commands: spicedb is the default (D-164) and needs an endpoint and, for the commands that can't wait, a preshared key from the environment or --spicedb-key-file, and serve starts without one, to wait for the file; plaintext gRPC comes from --spicedb-insecure or OPENFIEF_SPICEDB_INSECURE=true, as the core's pod sets it, and a value of that variable that is no boolean is refused; the policy, --authz=policy, needs nothing of SpiceDB and ignores its settings; a run with no flag and no SpiceDB endpoint, serve or token issue or revoke, fails at once with an error naming the endpoint's flag and variable and the opt-out --authz=policy; an unknown mode and a key given as a flag are refused; the key shows in no format verb, JSON or log line of the configuration; the policy wires identityapp.Policy and no syncer; spicedb wires its authorizer and syncer, and with nothing listening refuses every check with an error. A false pass would be a redaction that hides only one formatting, which the table of verbs and the JSON log would catch.