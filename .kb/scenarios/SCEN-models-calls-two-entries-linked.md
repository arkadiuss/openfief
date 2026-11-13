---
title: A call writes sent before it and its result after, under one call ID
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T14:45:40.985Z'
id: SCEN-models-calls-two-entries-linked
type: scenario
---
Given a working audit log, each model tried writes a model_call entry with the outcome sent before the call and a model_call_result entry with the same call ID after it, ok or failed with an error class; a fallback has its own pair under its own ID; a call refused by the switch or the data tiers has only a refused result entry, with no model; given a log that fails on the sent entry no model is called, and given one that fails on the result entry the answer is withheld.