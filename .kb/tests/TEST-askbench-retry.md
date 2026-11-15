---
title: A transport failure is retried once, a 429 once inside the chat client, any other 4xx or an answer never, and a host over five percent fails reliability
sourceFile: tools/askbench/retry_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:20.244Z'
status: active
id: TEST-askbench-retry
type: test
---
Runs the bench and the model's calls against a server that drops connections and answers 5xx and 4xx: a failed call is retried once and scored when the retry answers, two failures leave the question not scored, a 429 is retried inside the chat client and one that stays a 429 counts as a transport failure without a bench retry, any other 4xx or an answer is never retried, the retry spends the budget, the rate is transport-failed calls over calls made, the 5% threshold is exact, and the ranking leaves out an unreliable host.
