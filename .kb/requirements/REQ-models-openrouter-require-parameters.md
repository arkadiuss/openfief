---
title: On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools or a response format, and sends that member to no other server.
status: open
origin:
  kind: agent
  ref: EV1 brief from the author's M2 corrections and D-208, 2026-11-13
  recorded_at: '2026-11-13T20:00:55.400Z'
firmness: default
semantic_text: On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools or a response format, and sends that member to no other server.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 12258b587fd19ad691710c6ad9a717ad8b38b30513f3b7ef5eb2628a055575c6
semantic_inventory:
  - claim_key: CLAIM-00F5D8922837E9E9
    claim_text: On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools or a response format, and sends that member to no other server
    role: descriptive
    status: missing
    span:
      start: 0
      end: 167
logic_claims:
  - CLAIM-00F5D8922837E9E9
id: REQ-models-openrouter-require-parameters
type: req
---
On OpenRouter the model-openai plugin sets provider.require_parameters on every request that sends tools or a response format, and sends that member to no other server.

## Context

Without it OpenRouter may route a tool request to a host that drops tools and answers without a call, which the Ask box would present as an answer; a request for the structured answer needs it for the same reason, or a host that ignores the response format answers in prose. OpenAI's own API refuses unknown request members, so only OpenRouter gets it. The response format was added by the structured answer (AB3, 2026-11-14).

## Source

> EV1 brief from the author's M2 corrections and D-208, 2026-11-13; widened by The AB3 brief from the design session, 2026-11-14: (4) a structured final answer for the Ask box, with strict structured output where the model entry has the flag and JSON mode with one retry where it has not, validated cards and suggestions, the UI drawn from the fields, and the bench scoring by the schema where it decides.
