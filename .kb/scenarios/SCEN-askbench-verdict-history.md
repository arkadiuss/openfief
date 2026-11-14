---
title: A model@host's runs are kept only when asked and counted as consecutive passes or fails
origin:
  kind: agent
  ref: 'AB6 brief from the design session: retry and reliability in the bench, one prompt tweak (D-205, D-207, D-217, D-220), and the coordinator''s D-221 addition (history, weekly gate), 2026-11-14'
  recorded_at: '2026-11-14T19:56:43.431Z'
status: active
id: SCEN-askbench-verdict-history
type: scenario
---
Given runs of a model@host with the same configuration hash, when the bench is asked for a history with -history, then each whole run is added with its date, verdict, reliability and hash and the file is rewritten, no file is written without the flag, a run of some cases or one the budget stopped or one that could not run is not added, two consecutive passes (verdict and reliability both) make the label eligible for the route, a changed configuration hash starts the count again, and summary.md shows the consecutive passes or fails of every row.