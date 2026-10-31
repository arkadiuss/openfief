---
title: Only the core reaches an operator-deployed plugin
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:44.724Z'
id: SCEN-e2e-plugin-only-core-calls
type: scenario
tags:
  - acceptance:p10-17
---
Given the signed connector running, when the core syncs it with its own CLI the sync reports an empty batch, and when the suite's client pod or the outsider pod connects to the plugin's Service they get no connection.