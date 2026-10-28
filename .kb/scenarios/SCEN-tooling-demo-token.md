---
title: The API token is required and never shown
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:59:34.825Z'
id: SCEN-tooling-demo-token
type: scenario
---
Given HCLOUD_TOKEN is unset, when any demo command runs, then it fails naming HCLOUD_TOKEN without calling the API. Given a token, then no command line, output or written file of plan, up or down contains it.