---
title: An upgrade that requests a new host keeps the previous version running
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T06:36:19.461Z'
id: SCEN-plugins-rollout-upgrade-waits
type: scenario
---
Given the sample plugin running at its first digest under a grant for its two hosts, when its Plugin is changed to a second digest that also requests new.example.net, then the first digest keeps running with the granted hosts, and new.example.net:443 is reported as pending approval. Given no version ran before, then nothing runs until the grant approves it. Assumes the previous version is known from the operator's record.