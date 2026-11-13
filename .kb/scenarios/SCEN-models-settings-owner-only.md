---
title: Only the owner changes the models settings
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:15.902Z'
id: SCEN-models-settings-owner-only
type: scenario
---
Given a caller with no session, a read session, an owner session on another origin and the owner's session on an allowed origin, only the last may change the switch, a toggle, a route or a tier's models, and the plugin-facing service admits view plugins alone.