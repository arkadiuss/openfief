---
title: Each scorer of the bench is a goeval scorer, and only the bench imports goeval
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:29.691Z'
id: SCEN-askbench-scorers-implement-goeval
type: scenario
---
Given a run's answer, cards and tool calls, when the scorers grade it, then each is an api.Scorer, the tool calls are graded by goeval's tool-call scorer with the arguments the model wrote, and the checks that know OpenFief's data are the bench's own; a package other than the bench importing goeval, or the bench importing a package of it beyond api, heuristic and openai, is a violation of the architecture test.