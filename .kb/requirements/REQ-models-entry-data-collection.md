---
title: An entry of a model-openai instance at OpenRouter may set dataCollection to deny or allow, which the plugin sends as provider.data_collection merged into the same provider object as the pinned hosts and require_parameters, with that entry's requests only and only to OpenRouter; a config that sets it for another server or to another value is refused; and the demo's model entry sets deny, so a host that may store a request or train on it can never serve the demo.
status: open
firmness: default
semantic_text: An entry of a model-openai instance at OpenRouter may set dataCollection to deny or allow, which the plugin sends as provider.data_collection merged into the same provider object as the pinned hosts and require_parameters, with that entry's requests only and only to OpenRouter; a config that sets it for another server or to another value is refused; and the demo's model entry sets deny, so a host that may store a request or train on it can never serve the demo.
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T19:04:57.079Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3e87b712f424ce85c7d77749ef068ac95400e08dbc115f4f2c6311ebe27f6d3e
semantic_inventory:
  - claim_key: CLAIM-03B22C6296C75D7C
    claim_text: An entry of a model-openai instance at OpenRouter may set dataCollection to deny or allow, which the plugin sends as provider.data_collection merged into the same provider object as the pinned hosts and require_parameters, with that entry's requests only and only to OpenRouter
    role: descriptive
    status: missing
    span:
      start: 0
      end: 277
  - claim_key: CLAIM-A9156E0BD8370D12
    claim_text: a config that sets it for another server or to another value is refused
    role: descriptive
    status: missing
    span:
      start: 279
      end: 350
  - claim_key: CLAIM-9B74BF6E761C5A60
    claim_text: and the demo's model entry sets deny, so a host that may store a request or train on it can never serve the demo
    role: descriptive
    status: missing
    span:
      start: 352
      end: 464
logic_claims:
  - CLAIM-03B22C6296C75D7C
  - CLAIM-A9156E0BD8370D12
  - CLAIM-9B74BF6E761C5A60
id: REQ-models-entry-data-collection
type: req
---
An entry of a model-openai instance at OpenRouter may set dataCollection to deny or allow, which the plugin sends as provider.data_collection merged into the same provider object as the pinned hosts and require_parameters, with that entry's requests only and only to OpenRouter; a config that sets it for another server or to another value is refused; and the demo's model entry sets deny, so a host that may store a request or train on it can never serve the demo.

## Context

R-33 puts DeepInfra's terms (section 7(b), Zero Data Retention) at no retention and no training, and the demo's route is the one member openai/gpt-oss-120b on deepinfra/bf16 (D-207, D-221), but a pin can be edited and a fallback added later, and OpenRouter can route to hosts whose terms are worse. data_collection: deny makes OpenRouter itself refuse any host that may store or train on a request, whatever the pin says. It is OpenRouter's request member, so it is sent only to OpenRouter and refused in the config of another server, like hosts and reasoning (REQ-models-entry-hosts-reasoning).

## Source

> The design session's M5 corrections: send OpenRouter provider.data_collection: deny on the demo's route/entry so a host that may store or train can never serve it; a per-entry dataCollection: deny option in model-openai (OpenRouter only, validated, merged into the same provider object as order/allow_fallbacks/require_parameters), set on the demo's gemma entry, with a table test on the body (M5 brief, 2026-11-14).
