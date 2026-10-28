---
title: A pod no policy admits cannot reach the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:01:00.009Z'
id: SCEN-e2e-core-policy-enforced
type: scenario
---
Given the core available and admitted traffic from the client pod answered, when a pod in the same namespace without the admitted label calls the core's Service, then it gets no connection, which shows the cluster's CNI enforces NetworkPolicy as D-12 requires.