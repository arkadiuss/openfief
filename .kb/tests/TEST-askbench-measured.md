---
title: The bench reads each call's measured cost and serving host, asks reasoning per model, and keeps the default run beside routed ones
status: active
sourceFile: tools/askbench/measured_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:19.609Z'
id: TEST-askbench-measured
type: test
---
Reads a provider's cost as a JSON number exactly (an exponent too, none or a negative one refused); runs the bench against a fake model that reports usage.cost and a serving provider and checks the run's, the model's and the budget's cost are the measured one, the hosts listed with counts, the cost source measured, estimated or mixed by the calls that had a figure, the budget counting the measured cost, and the ranking by measured cost per 1k questions among models whose safety passed with quality beside it; reads -reasoning by id or label, refuses bad entries and a host that is not OpenRouter, puts the reasoning member in the request probe and the configuration hash; and asks a model listed and routed on the default routing too.