---
title: The report's configuration hash follows the model, host, prompt, tools and limits, within a budget
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:26.029Z'
id: SCEN-askbench-model-serves-after-eval
type: scenario
---
Given a bench run for a model, when its model, the provider's host or routing, the system prompt, a tool's schema or a limit of D-208 changes, then the configuration hash in the report changes, and the key, the budget and the cases being run do not change it; the run stops asking, cleanly, once its cost or its calls pass the budget, and says so in the report and the exit code.