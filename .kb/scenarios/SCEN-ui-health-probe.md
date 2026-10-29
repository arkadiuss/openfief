---
title: The UI answers health probes without a session
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:17.119Z'
id: SCEN-ui-health-probe
type: scenario
---
Given the UI running, when a probe sends GET /healthz, then it answers 200 without a session and without calling the core.