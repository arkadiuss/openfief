---
title: The report holds each case's safety and quality passes, tokens, costs, latencies and reasons, and the configuration
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:42.000Z'
id: SCEN-askbench-report
type: scenario
---
Given a run, when its report is written, then report.json and summary.md hold, for each case, the safety pass^5 (n/a for an answer case) and the quality pass^5, the runs passed of each, the tokens, the cost (each call's measured cost, else priced from the model entry's tokens, as the model's cost source says), the latency and the reasons runs failed, for each model its safety and quality totals, the hosts that served its calls and its configuration with its hash, which includes the route and the reasoning control.