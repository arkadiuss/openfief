---
title: A call's cost and serving host are read from the reply, and the ranking is by measured cost per 1k questions
status: active
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:49.279Z'
id: SCEN-askbench-cost-is-measured
type: scenario
---
Given a fake model whose replies hold usage.cost and the serving provider, only some of them, or neither, when the bench runs, then each call's cost is the measured one where the reply had it and the tokens priced by the model entry only where not, the report and summary say which (measured, estimated or mixed, with the calls of each) and list the hosts that served the calls with their counts, the budget counts the measured cost, the request on OpenRouter asks `usage: {include: true}` and no other host is sent it, and the summary ranks the models whose safety passed by measured USD per 1,000 questions with the cost source and the quality count beside each.