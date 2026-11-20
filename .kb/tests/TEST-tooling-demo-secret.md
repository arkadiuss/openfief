---
title: secret.sh sends the value on stdin only and refuses what it must
status: active
sourceFile: scripts/tests/demo-secret.test.sh
verification_scope: integration
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data, 2026-11-19'
  recorded_at: '2026-11-19T16:24:01.275Z'
id: TEST-tooling-demo-secret
type: test
---
Runs scripts/demo/secret.sh with a fake ssh that logs its arguments and keeps its standard input, and a fake hcloud: the key from the environment or a pipe reaches ssh on stdin exactly, with no newline, and is in no argument, log or output; the remote command is exactly the create-or-replace; the Turnstile command names its own Secret; rotation sends the new value; and each refusal (no DEMO_SSH=keys, no value, white space, no token, bad ENV, unknown machine, unknown secret) sends nothing, with a failing ssh failing the task. The grafana command sends the endpoint, instance ID and token, from the environment or three piped lines, as one env file on stdin into the create-or-replace of grafana-otlp in observability followed by the collector's restart, drops an endpoint's trailing slash, and prints and passes in no argument any of the three values; an http endpoint, one on another host, one that only starts like grafana.net, one with credentials in it, an instance ID that is no number, a token with white space or too short, a missing token and a missing DEMO_SSH=keys each send nothing.