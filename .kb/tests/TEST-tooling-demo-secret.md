---
title: secret.sh sends the value on stdin only and refuses what it must
status: active
sourceFile: scripts/tests/demo-secret.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-14T19:05:45.836Z'
id: TEST-tooling-demo-secret
type: test
---
Runs scripts/demo/secret.sh with a fake ssh that logs its arguments and keeps its standard input, and a fake hcloud: the key from the environment or a pipe reaches ssh on stdin exactly, with no newline, and is in no argument, log or output; the remote command is exactly the create-or-replace; the Turnstile command names its own Secret; rotation sends the new value; and each refusal (no DEMO_SSH=keys, no value, white space, no token, bad ENV, unknown machine, unknown secret) sends nothing, with a failing ssh failing the task.