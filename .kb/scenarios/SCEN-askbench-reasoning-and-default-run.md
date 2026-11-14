---
title: A model's reasoning control is sent on OpenRouter only and a listed model keeps its default-routing run beside its routes
status: active
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:50.227Z'
id: SCEN-askbench-reasoning-and-default-run
type: scenario
---
Given -reasoning entries by model id or label, when the bench builds its requests, then OpenRouter gets `reasoning: {enabled: false}` for off and `reasoning: {effort: low|medium|high}` for the others and nothing when a model has no entry, another host is refused as a usage error and never sent it, the setting is in the configuration hash, bad entries are usage errors, and given a model that -model lists and -route pins, then it runs once on the default routing and once on each route, while a model only a route names runs on its routes.