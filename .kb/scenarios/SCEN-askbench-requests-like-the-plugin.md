---
title: The bench's requests are the plugin's
status: active
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:04.229Z'
id: SCEN-askbench-requests-like-the-plugin
type: scenario
---
Given a fake OpenAI-compatible server, when the bench asks it, then the request is the plugin's chat client's, the key is in the Authorization header and in no report, summary, log or error, even when the provider fails or quotes it back; and on OpenRouter's host a request with tools or a response format carries provider.require_parameters.