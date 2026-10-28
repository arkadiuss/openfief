---
title: The UI's code reaches the record only through the read API
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:19.018Z'
id: SCEN-ui-imports
type: scenario
---
Given the UI's Go package, when its imports are listed, then none is an OpenFief package other than the kernel and the SDK's generated clients and helpers.