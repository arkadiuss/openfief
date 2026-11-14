---
title: The Ask box's structured answer is strict where the entry has the flag and JSON mode where it has not
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:52:58.634Z'
id: SCEN-models-structured-answer-path
type: scenario
---
Given an Ask call with a response schema, when the router plans it, then a model whose entry has no structured output still serves it, and the provider's request says JSON mode is allowed; the plugin sends strict json_schema for an entry with the flag and json_object for one without, refuses a strict request for an entry without it, and on OpenRouter asks require_parameters for either; a schema that is no JSON or is too long is refused before any model is called.