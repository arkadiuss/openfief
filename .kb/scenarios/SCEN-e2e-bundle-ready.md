---
title: The operator, egress proxy and core installed from the bundle become available
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T13:50:37.305Z'
id: SCEN-e2e-bundle-ready
type: scenario
---
Given the bundle's CRDs, a CloudNativePG cluster openfief-db and the bundle's install applied to an empty namespace, when the suite waits for each Deployment in the bundle, then the operator, the egress proxy and the core each become available within five minutes; a pod that cannot start fails the wait at once.