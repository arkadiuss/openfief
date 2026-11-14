---
title: The Ask box bench sends its requests to the model through the model-openai plugin's own chat client, so a request carries provider.require_parameters on OpenRouter and the call's output limit as max_tokens, and the provider key goes only to the provider and is in no report, log or error.
status: open
origin:
  kind: agent
  ref: 'AB1 brief from the author: the Ask box bench (D-217, D-218, D-219), 2026-11-14'
  recorded_at: '2026-11-14T08:23:38.617Z'
firmness: default
semantic_text: The Ask box bench sends its requests to the model through the model-openai plugin's own chat client, so a request carries provider.require_parameters on OpenRouter and the call's output limit as max_tokens, and the provider key goes only to the provider and is in no report, log or error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 84e9a1d0fe88e58f4c4cf354cec02df92a2409b8a93a0af35389bddfd68bd735
semantic_inventory:
  - claim_key: CLAIM-CCF83433ED3D5F47
    claim_text: The Ask box bench sends its requests to the model through the model-openai plugin's own chat client, so a request carries provider.require_parameters on OpenRouter and the call's output limit as max_tokens, and the provider key goes only to the provider and is in no report, log or error
    role: descriptive
    status: missing
    span:
      start: 0
      end: 287
logic_claims:
  - CLAIM-CCF83433ED3D5F47
id: REQ-askbench-requests-like-the-plugin
type: req
---
The Ask box bench sends its requests to the model through the model-openai plugin's own chat client, so a request carries provider.require_parameters on OpenRouter and the call's output limit as max_tokens, and the provider key goes only to the provider and is in no report, log or error.

## Context

The author asked that the bench call the model the way production does, with the key of the model router's OpenRouter provider, and that no output ever hold it (GR-12).

## Source

> The author's brief for the Ask box bench (AB1), 2026-11-14, on D-217, D-218 and D-219.
