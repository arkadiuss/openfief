---
title: openfief enroll prints the one-time link for an allowed origin, or forwards to the core and prints it for localhost
status: active
sourceFile: internal/identity/cli/enroll_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:50.401Z'
id: TEST-identity-cli
type: test
---
Runs identity/cli.Run over a fake core. Without --origin: with the defaults it forwards port 8080, asks for the link for localhost:8080 with 15 minutes, prints the link, the first-run wizard and Ctrl-C, and stops the forward when the context ends; another port warns about --auth-origin; bad flags, no cluster, a failed forward and a failed link print no link. With --origin (enroll_origin_test.go): for an origin among the core's allowed origins it asks for the link for that origin, forwards no port and prints the link and the wizard without any port-forward wording; an origin the core doesn't allow is ErrOriginNotAllowed, --origin with --port and a host without a scheme are usage errors, and a core whose origins can't be read fails, each before any link.