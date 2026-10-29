---
title: openfief enroll forwards to the core, prints the one-time link for that origin and keeps forwarding
status: active
sourceFile: internal/identity/cli/enroll_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:50.401Z'
id: TEST-identity-cli
type: test
---
Runs identity/cli.Run over a fake core: with the defaults it forwards port 8080, asks for the link for localhost:8080 with 15 minutes, prints the link, the first-run wizard and Ctrl-C, and stops the forward when the context ends; another port warns about --auth-origin; bad flags, no cluster, a failed forward and a failed link fail without printing a link.