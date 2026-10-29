---
title: Deleting a Plugin removes its workload and allow-list entry
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:49.992Z'
id: SCEN-e2e-plugin-uninstalled
type: scenario
---
Given the signed plugin running, when its Plugin is deleted, then its Deployment, Service and NetworkPolicy are removed and its key disappears from openfief-egress-allowlist within two minutes.