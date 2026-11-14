---
title: The bench puts the plugin's request on the wire and the key nowhere else
status: active
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:24:38.575Z'
id: SCEN-askbench-requests-like-the-plugin
type: scenario
---
Given a fake OpenAI-compatible server, when the bench asks it, then the request is the plugin's chat client's, the key is in the Authorization header and in no report, summary, log or error, even when the provider fails or quotes it back; and on OpenRouter's host a request with tools carries provider.require_parameters.