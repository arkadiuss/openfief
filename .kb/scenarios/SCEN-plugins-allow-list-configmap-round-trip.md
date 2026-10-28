---
title: The operator's allow-list ConfigMap is read back exactly
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:50.195Z'
id: SCEN-plugins-allow-list-configmap-round-trip
type: scenario
---
Given allow-lists for plugins sample (api.example.com:443 and data.example.org:8443) and xtb (no hosts), when they are encoded, then the ConfigMap data has key sample with two sorted host:port lines and key xtb with an empty value, and decoding that data returns the same lists; data with a host written without its port is refused.