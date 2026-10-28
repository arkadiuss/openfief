---
title: The built-in UI's pod becomes ready on a local cluster
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:20:57.431Z'
id: SCEN-e2e-ui-ready
type: scenario
---
Given the demo's stack installed on k3d or kind from this checkout's images, when the UI's Deployment plugin-ui rolls out, then it becomes available with a ready pod, answering its health endpoint.