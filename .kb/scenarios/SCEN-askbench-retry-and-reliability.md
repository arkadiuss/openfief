---
title: A transport failure is retried once and a host that fails too often fails reliability
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:42.363Z'
status: active
id: SCEN-askbench-retry-and-reliability
type: scenario
---
Given a model whose host drops a call, answers 5xx, answers 4xx or answers, when the bench asks, then a call that failed with no answer, a timeout or a 5xx is retried once as a call of its own (the budget and the cost count it) and the run's detail marks it retried and whether the retry answered, a retry that answers leaves the question scored, a failed retry leaves it not scored and listed once, a 4xx and any answer are never retried; and per model@host the transport-failure rate is the transport-failed calls over the calls made, a rate above 5% fails reliability for the run whatever the verdict, report.json and summary.md show the rate and pass or fail beside the verdict, and the ranking lists only models whose verdict and reliability pass.