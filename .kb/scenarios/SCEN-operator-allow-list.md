---
title: The allow-list ConfigMap follows what runs
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:59:58.437Z'
id: SCEN-operator-allow-list
type: scenario
---
Given a granted Plugin with egress host api.example.com, when it runs, then the openfief-egress-allowlist ConfigMap has its key with api.example.com:443. When its grant is revoked or the Plugin is deleted, then its key is removed. Assumes other plugins' keys are left to their own reconciles.