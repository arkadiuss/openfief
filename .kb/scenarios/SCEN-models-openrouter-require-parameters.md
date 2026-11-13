---
title: Only OpenRouter, and only with tools, gets require_parameters
status: active
origin:
  kind: agent
  ref: EV1 brief from the author's M2 corrections and D-208, 2026-11-13
  recorded_at: '2026-11-13T20:00:50.702Z'
id: SCEN-models-openrouter-require-parameters
type: scenario
---
Given a request that sends tools, when the base URL's host is openrouter.ai, then the body holds provider.require_parameters true; with no tools, or at any other host, the body has no provider member.