---
title: A remote context is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-27T20:33:09.742Z'
id: SCEN-tooling-e2e-remote-context
type: scenario
---
Given the current kubeconfig context points at https://cluster.example.com:6443, when task e2e runs, then it refuses with a message naming the context and exits non-zero without running tests.
