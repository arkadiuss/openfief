---
title: The figures go through the router as T0, the instructions as T2, with a strict schema
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:27.680Z'
id: SCEN-maintainer-sends-figures-through-router
type: scenario
---
Given a route of one model, when the maintainer asks, then the request has the instructions as public data and the weights as whole-portfolio data, a strict schema and no tools; a model not allowed for both tiers or without structured output is a refusal, and the audit entries name the maintainer role.