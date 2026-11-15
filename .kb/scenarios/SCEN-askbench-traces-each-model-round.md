---
title: A run's report entry traces each call to the model, so a host that reports only uncached tokens is told from a prompt that did not arrive whole
status: active
origin:
  kind: agent
  ref: 'Live bench finding on R1 (gpt-oss-120b on deepinfra: two model calls, 1694 input tokens, a greeting), 2026-11-14'
  recorded_at: '2026-11-14T23:21:33.246Z'
id: SCEN-askbench-traces-each-model-round
type: scenario
---
Given a case asked of a model whose first reply is plain words and whose second is the structured answer, then report.json holds two rounds for each run: the first with roles system and user, the second with system, user, assistant and user, each message with its length and a short hash and none with its text, and each with the prompt and completion tokens the host reported, its finish reason and its reply's text; given a reply that reports cached tokens and asks for a tool, then the round holds the cached count and the tool's name, and a reply's text over 500 characters is cut to 500; and the OpenAI-compatible client reads cached tokens from usage.prompt_tokens_details.cached_tokens or the usage's cached_tokens.