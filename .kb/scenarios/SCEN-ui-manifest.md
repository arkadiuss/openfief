---
title: The ui plugin asks for the view capability only
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:23.317Z'
id: SCEN-ui-manifest
type: scenario
---
Given plugins/ui/manifest.json, when the operator's label parser reads it, then it holds only the view capability, no egress host and no Secret, with a config schema.