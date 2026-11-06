---
title: 'Attempts to bypass the egress proxy fail: no direct connection and no DNS lookup'
status: active
tags:
  - acceptance:p10-10
  - acceptance:p10-16
origin:
  kind: agent
  recorded_at: '2026-11-06T20:30:00.053Z'
id: SCEN-e2e-egress-bypass-fails
type: scenario
---
Given the nbp-fx connector granted an egress host, whose pod reaches the egress proxy, when from inside its pod it connects directly to the Kubernetes API server own address and to a TEST-NET address, and looks up an outside name and a cluster name, then every connection fails and every name lookup fails, while a pod without a policy reaches that API server address and resolves that cluster name. Runs on both kind and k3d.