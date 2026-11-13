---
title: On kind and k3d, with the core on SpiceDB, a script token granted one sleeve views that sleeve's accounts alone, a later account is covered, a revoked grant views nothing and the log names each grant by ID
status: active
sourceFile: test/e2e/grants_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-13T10:00:11.447Z'
id: TEST-e2e-grants
type: test
---
Runs against the real stack whose core answers from SpiceDB (the suite's overlay). The owner's session is a virtual authenticator's passkey made through an enrollment link. The UI plugin, installed before the core started, lists every demo account with its own token: the cutover gave it the household grant. A script token with no grant lists nothing and its account calls answer 404. Granted the defensive sleeve it lists that sleeve's account alone, and its allocation by sleeve holds that sleeve alone while a household token's holds both; an account created in the sleeve afterwards is listed too, every account outside the sleeve answers 404; revoked, the next call lists nothing. Granted one account it lists that one alone. The owner reads the audit log: grant_created and grant_revoked name the token and the target by kind and ID and nothing else.