---
title: core init-credentials needs no database, takes no argument, creates the two Secrets and logs their names and outcomes, never a value
status: active
sourceFile: cmd/core/initcredentials_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:46.251Z'
id: TEST-core-init-credentials
type: test
---
Parses the command line (TestParseArgs: init-credentials needs no DSN, ignores one, and refuses an argument or a flag) and runs initCredentials over the fake clientset with the system's random source and a JSON log. Every call to the API server is a create of a Secret in openfief and the two created are the preshared key's and the role's; with the key's Secret fed beforehand the log says it was kept and the owner's value is unchanged; the log names each Secret with created or kept and holds no value of either Secret.