---
title: The CLI lists installed instances by name
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T22:43:27.010Z'
id: SCEN-plugins-cli-list-instances
type: scenario
---
Given installed instances unrelated, sample-b and sample-a, when the owner lists plugins, then they are listed as sample-a, sample-b, unrelated, each with its image, whether its Plugin and grant exist, its signature status, readiness and whether permissions are pending.