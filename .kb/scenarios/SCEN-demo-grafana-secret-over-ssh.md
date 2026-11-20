---
title: task demo:grafana-secret sends the three Grafana values on stdin only and refuses what it must
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:43.074Z'
id: SCEN-demo-grafana-secret-over-ssh
type: scenario
---
Given a fake ssh and a fake hcloud, when task demo:grafana-secret runs, then the endpoint, instance ID and token from the environment or a pipe reach ssh on standard input as an env file and are in no argument, log or output, the remote command creates or replaces grafana-otlp in observability and restarts the collector, a trailing slash on the endpoint is dropped, and an http endpoint, one on another host, one that only starts like grafana.net, one with credentials in it, an instance ID that is no number, a token with white space or too short, a missing token and a missing DEMO_SSH=keys each send nothing and print no value.