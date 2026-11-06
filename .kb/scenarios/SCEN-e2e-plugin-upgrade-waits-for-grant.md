---
title: An upgrade that asks for a new egress host keeps the old version running until its grant approves the host
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:37.793Z'
id: SCEN-e2e-plugin-upgrade-waits-for-grant
type: scenario
---
Given the synthetic connector running under a grant for exactly what its first image requests, when the owner upgrades its Plugin to the second image, whose signed manifest asks for example.org too, then the operator keeps the first image deployed and running, reports Ready True with reason RunningPreviousVersion and example.org:443 pending approval, and keeps the host off the allow-list; when the grant then approves example.org, the second image runs with the host on the allow-list and nothing pending. Runs on both kind and k3d.