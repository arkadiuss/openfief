---
title: Any reader learns the reporting currency, the number format and the wrappers
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T00:04:56.895Z'
id: SCEN-views-settings
type: scenario
---
Given the Polish pack and no chosen format, when GET /v1/settings is read, then it answers PLN, space-comma as the default and the four wrappers; after the owner chooses comma-dot, that, not the default.