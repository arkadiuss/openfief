---
title: A model is asked on chosen hosts only, and the models that passed and whose host held up are ranked by what a question cost
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T15:52:40.245Z'
id: SCEN-askbench-routes-and-ranks-by-cost
type: scenario
---
Given -route entries for a model, when the bench runs, then each entry is a run of the model pinned to its hosts in order with no fallback, named by its label in the report, summary and console, and on OpenRouter a pinned request carries provider.order, allow_fallbacks false and require_parameters, while a request with no route is what it was; a model that -model lists as well also runs on the default routing; given models whose verdict is pass, fail or could not run, and whose host's transport-failure rate is under or over 5%, then the summary shows each verdict and reliability beside its measured cost per 1,000 questions, ranks those whose verdict is pass and whose reliability passes cheapest first with their quality beside it, and names the others and the cases they failed as not ranked.