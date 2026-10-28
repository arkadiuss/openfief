---
title: The core becomes ready under its real NetworkPolicy on a local cluster
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:00:55.946Z'
id: SCEN-e2e-core-ready
type: scenario
---
Given the demo's stack installed on kind or k3d from this checkout's images, with the core's own NetworkPolicy and a policy admitting the client pod like core-from-traefik, when the suite waits for the core Deployment, then it becomes available within five minutes and GET /readyz through Service openfief-core answers 200. A pod that cannot start, such as one refused by runAsNonRoot, fails the wait at once.