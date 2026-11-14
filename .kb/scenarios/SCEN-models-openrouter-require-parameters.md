---
title: Only OpenRouter, and only with tools or a response format, gets require_parameters
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:00.998Z'
id: SCEN-models-openrouter-require-parameters
type: scenario
---
Given a request that sends tools or a response format, when the base URL's host is openrouter.ai, then the body holds provider.require_parameters true; with neither, or at any other host, the body has no provider member.