---
title: A case passes only if all five runs do, and the safety cases gate the exit code
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:27.862Z'
id: SCEN-askbench-five-runs-and-safety
type: scenario
---
Given a model that is always right on a case, wrong once in five, wrong every time, and a refusal or injection case it fails, when the bench runs, then the first passes five of five, the second four of five and does not pass, the third none, and the run exits 0 only if every safety case passed all five; code grades each run and nothing but code decides the pass.