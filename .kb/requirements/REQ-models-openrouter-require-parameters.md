---
title: On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools, and sends that member to no other server.
status: open
origin:
  kind: agent
  ref: EV1 brief from the author's M2 corrections and D-208, 2026-11-13
  recorded_at: '2026-11-13T20:00:55.400Z'
firmness: default
semantic_text: On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools, and sends that member to no other server.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3821f668d41fb9c9e797dda37dcadef05a0bb06de763f2e528db71bdac44294f
semantic_inventory:
  - claim_key: CLAIM-369622DC47793449
    claim_text: On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools, and sends that member to no other server
    role: descriptive
    status: missing
    span:
      start: 0
      end: 146
logic_claims:
  - CLAIM-369622DC47793449
id: REQ-models-openrouter-require-parameters
type: req
---
On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools, and sends that member to no other server.

## Context

Without it OpenRouter may route a tool request to a host that drops tools and answers without a call, which the Ask box would present as an answer. OpenAI's own API refuses unknown request members, so only OpenRouter gets it.

## Source

> EV1 brief from the author's M2 corrections and D-208, 2026-11-13