---
title: The XTB parser runs in its own pod and has no egress at all
status: active
tags:
  - acceptance:p10-10
origin:
  kind: agent
  recorded_at: '2026-11-06T20:29:58.668Z'
id: SCEN-e2e-xtb-no-egress
type: scenario
---
Given the XTB connector installed with its own manifest, which requests no egress host, and a grant that approves only its capability and institution, when the operator deploys it, then it runs alone in a pod of its own, its NetworkPolicy allows egress only to the core pods on the core plugin port, it gets no HTTPS_PROXY, no egress token, no effective egress host and an empty allow-list entry, and from inside its pod the core plugin port answers while the egress proxy, the core browser port, another plugin and the Kubernetes API server own address get no connection and name lookups fail, although a pod without a policy reaches that address and resolves that name. Runs on both kind and k3d.