---
title: A verdict is shown with the version of the scoring rules that graded it
status: active
origin:
  kind: agent
  ref: 'Design session decisions (author''s delegate): a scorer version and offline re-scoring of the Ask box bench, with D-217, D-219, D-220 and D-221, 2026-11-15'
  recorded_at: '2026-11-15T15:38:48.250Z'
id: SCEN-askbench-scorer-version
type: scenario
---
Given a bench run, a re-graded report and result files written before versioning, when the bench writes report.json and summary.md, docgen imports a report and the results page is rendered, then report.json, each trimmed result file and each history entry carry the scorer version, the summary and the page show it beside every verdict, a file without one reads as scorer 1 before versioning, and the configuration hash is the same whatever the scorer version.