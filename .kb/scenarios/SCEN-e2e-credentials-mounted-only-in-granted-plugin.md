---
title: A connector's credentials Secret is mounted only into the pod of the instance granted it
status: active
tags:
  - acceptance:p10-14
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:42.113Z'
id: SCEN-e2e-credentials-mounted-only-in-granted-plugin
type: scenario
---
Given the synthetic connector instance e2e-books running under a grant that names its Secret e2e-books--credentials, next to the core, the operator, the egress proxy and other plugins, when every pod in the openfief namespace is read, then exactly the instance own pods use that Secret, as a volume, a projected source or an environment variable, and no other pod, the core included, does. Runs on both kind and k3d.