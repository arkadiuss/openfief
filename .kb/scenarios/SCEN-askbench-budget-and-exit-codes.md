---
title: The bench aborts cleanly past its budget and exits by its safety cases
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:40.322Z'
id: SCEN-askbench-budget-and-exit-codes
type: scenario
---
Given a budget of calls or of cost too small for the run, when the bench runs, then it stops asking, marks the questions not asked, writes its report and exits 3; with every safety case passing on all runs it exits 0, with one failing it exits 1, and without the key it makes no call and exits 2.