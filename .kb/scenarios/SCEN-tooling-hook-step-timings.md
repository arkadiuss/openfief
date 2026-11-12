---
title: Each hook step says how long it took
status: active
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:43.865Z'
id: SCEN-tooling-hook-step-timings
type: scenario
---
Given a plan of two steps, when the hook runs them, then each step's line says its seconds and the hook ends with the total; when a step fails, that step's time and the total are still printed and the steps after it do not run.
