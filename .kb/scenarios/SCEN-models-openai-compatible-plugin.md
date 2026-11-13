---
title: The plugin speaks the OpenAI-compatible chat protocol to its configured provider only
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T11:28:18.641Z'
id: SCEN-models-openai-compatible-plugin
type: scenario
---
Given a fake OpenAI-compatible server, the plugin lists the configured entries, describes where they run and the retention terms, sends chat completions with the key from its Secret, tools, structured output and an output limit, maps the usage and finish reason back, classifies a refused key, a rate limit or a failing provider as unavailable and a rejected request as invalid, and never quotes the request or the key in an error; its manifest asks for the model-provider capability and the host of its base URL only.