---
title: The telemetry collector reaches only Grafana's exporter, with its credentials from a Secret that is not in git, and waits without it
status: active
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:42.533Z'
id: SCEN-demo-telemetry-egress
type: scenario
---
Given the demo's telemetry layer, when the demo is validated, then the collector has the one exporter otlphttp/grafana taking its endpoint, instance ID and token from the environment the Secret grafana-otlp fills (three required keys), the build holds no Secret and the configuration no Grafana address, the NetworkPolicy admits the core and the operator on 4318 and lets the collector out to DNS and HTTPS with the private ranges left out, the node's pod logs are mounted read-only, the image is pinned by tag and digest, and the live demo sends OTLP to the collector's Service; a change to any of these fails naming the problem.