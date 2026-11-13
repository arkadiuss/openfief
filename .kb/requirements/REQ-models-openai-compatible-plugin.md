---
title: 'The first model-provider plugin speaks the OpenAI-compatible chat API, configured in three layers: a wire protocol, a provider instance with its base URL, the Secret holding its key and its timeout, and model entries with id, price per token, limits and flags for tool calling and structured output.'
status: open
text_ref: design-decisions:D-205#1
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-205#1
  recorded_at: '2026-11-13T11:28:52.570Z'
semantic_text: 'The first model-provider plugin speaks the OpenAI-compatible chat API, configured in three layers: a wire protocol, a provider instance with its base URL, the Secret holding its key and its timeout, and model entries with id, price per token, limits and flags for tool calling and structured output.'
semantic_source_field: semantic_text
semantic_source_hash: fe21b10435ca928d10f89b33f0ca76d13867898e73044ce97ebf083b336b7a96
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-71C0A1F0C0DC26CD
    claim_text: 'The first model-provider plugin speaks the OpenAI-compatible chat API, configured in three layers: a wire protocol, a provider instance with its base URL, the Secret holding its key and its timeout, and model entries with id, price per token, limits and flags for tool calling and structured output'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 298
logic_claims:
  - CLAIM-71C0A1F0C0DC26CD
semantic_clauses:
  - 'The first model-provider plugin speaks the OpenAI-compatible chat API, configured in three layers: a wire protocol, a provider instance with its base URL, the Secret holding its key and its timeout, and model entries with id, price per token, limits and flags for tool calling and structured output'
required_evidence: asserted
id: REQ-models-openai-compatible-plugin
type: req
---
The first model-provider plugin speaks the OpenAI-compatible chat API, configured in three layers: a wire protocol, a provider instance with its base URL, the Secret holding its key and its timeout, and model entries with id, price per token, limits and flags for tool calling and structured output.

## Context

plugins/model-openai is that plugin; one image is installed once per provider instance, so OpenAI, OpenRouter, Ollama, vLLM and LM Studio are each an instance with their own grant.

## Source

> Models are reached through model-provider plugins, configured in three layers as OpenClaw does: a wire protocol (OpenAI-compatible chat first, Anthropic's and Gemini's later), a provider instance (base URL, the Secret holding its key, timeout) and model entries (id, price per token, context and output limits, and flags for tool calling and structured output). [...] The first plugin speaks the OpenAI-compatible API, which serves OpenAI, OpenRouter, Ollama, vLLM, LM Studio and others.

Source: OpenFief author decisions (DESIGN.md section 10) - D-205#1