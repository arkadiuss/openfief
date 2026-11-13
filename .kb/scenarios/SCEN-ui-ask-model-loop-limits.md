---
title: A question stops at its rounds, tool calls and output tokens
status: active
origin:
  kind: agent
  ref: EV1 brief from the author's M2 corrections and D-208, 2026-11-13
  recorded_at: '2026-11-13T20:00:49.706Z'
id: SCEN-ui-ask-model-loop-limits
type: scenario
---
Given a model that keeps calling tools, when one question runs, then at most five rounds offer tools and one last offers none, at most ten tools run and the calls beyond get a refusal, each call is limited to the output tokens the question has left, no call is made once the 800 are spent, and a limit set in the configuration replaces its default.