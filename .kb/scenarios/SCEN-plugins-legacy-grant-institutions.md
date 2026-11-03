---
title: A grant made before D-185 keeps its plugin booking
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:26.851Z'
id: SCEN-plugins-legacy-grant-institutions
type: scenario
---
Given a PluginGrant without an institutions field, as every grant made before D-185, and a connector version running under it, when the operator reconciles, then the grant approves the institutions of that running version, which keeps running; a new version declaring another institution waits for the owner's approval while the old one runs; with nothing running, the grant approves none.