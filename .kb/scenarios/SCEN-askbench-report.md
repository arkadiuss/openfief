---
title: The report holds each case's passes, tokens, costs, latencies and reasons, and the configuration
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:42.000Z'
id: SCEN-askbench-report
type: scenario
---
Given a run, when its report is written, then report.json and summary.md hold, for each case, pass^5, the runs passed, the tokens, the cost from the model entry's prices, the latency and the reasons runs failed, and for each model the configuration with its hash.