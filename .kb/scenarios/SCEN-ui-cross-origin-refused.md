---
title: State-changing requests from another origin are refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:01.052Z'
id: SCEN-ui-cross-origin-refused
type: scenario
---
Given a POST whose Sec-Fetch-Site is not same-origin, or without Sec-Fetch-Site and with an Origin other than the UI's, or with neither header, when it reaches the UI, then it gets 403 before any handler runs; a same-origin POST passes.