---
title: A run no model answered could not run, and a failed call is an error with its status
status: active
origin:
  kind: agent
  ref: Brief to the Ask bench status session (a run no model answered must not look like a model result), 2026-11-14
  recorded_at: '2026-11-14T13:26:23.089Z'
id: SCEN-askbench-errors-not-model-failures
type: scenario
---
Given a provider that answers every call 402, 401, 429 or 500, when the bench runs, then it exits 2, says on stdout and in summary.md that the model could not run, and lists the calls by status and label (402 no credits x10) with no safety failure; given a provider that fails only some calls, the answered questions are scored, the failed ones are listed as errors and not scored, and neither the provider's reply nor the key is in any output.
