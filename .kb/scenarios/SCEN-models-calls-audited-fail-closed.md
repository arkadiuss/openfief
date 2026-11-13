---
title: A call is recorded before it is sent and refused when it cannot be recorded
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:20.551Z'
id: SCEN-models-calls-audited-fail-closed
type: scenario
---
Given an audit log that works, each call to a model and each refusal becomes a model_call entry naming the actor, the role and the model, with an outcome and no text; given a log that fails, no model is called and the error says the entry could not be written.