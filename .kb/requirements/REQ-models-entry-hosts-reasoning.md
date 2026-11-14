---
title: An entry of a model-openai instance at OpenRouter may pin its model to hosts, in order and with no fallback to others, and set its reasoning to off, low, medium or high; the plugin sends each only with that entry's requests and only to OpenRouter, and refuses a config that sets one for another server.
status: open
firmness: default
semantic_text: An entry of a model-openai instance at OpenRouter may pin its model to hosts, in order and with no fallback to others, and set its reasoning to off, low, medium or high; the plugin sends each only with that entry's requests and only to OpenRouter, and refuses a config that sets one for another server.
origin:
  kind: agent
  ref: 'M4 prep brief from the design session: the demo''s model behind the switch (D-207 amended, D-205, D-208, D-212, D-203), 2026-11-14'
  recorded_at: '2026-11-14T17:59:13.666Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fade5624526c4dbc00b67c35cd43a98fa8aec11083878c4051037b2d308462ff
semantic_inventory:
  - claim_key: CLAIM-6FEA4BE5A6EEA67B
    claim_text: An entry of a model-openai instance at OpenRouter may pin its model to hosts, in order and with no fallback to others, and set its reasoning to off, low, medium or high
    role: descriptive
    status: missing
    span:
      start: 0
      end: 168
  - claim_key: CLAIM-E66A51B216A49AC9
    claim_text: the plugin sends each only with that entry's requests and only to OpenRouter, and refuses a config that sets one for another server
    role: descriptive
    status: missing
    span:
      start: 170
      end: 301
logic_claims:
  - CLAIM-6FEA4BE5A6EEA67B
  - CLAIM-E66A51B216A49AC9
id: REQ-models-entry-hosts-reasoning
type: req
---
An entry of a model-openai instance at OpenRouter may pin its model to hosts, in order and with no fallback to others, and set its reasoning to off, low, medium or high; the plugin sends each only with that entry's requests and only to OpenRouter, and refuses a config that sets one for another server.

## Context

The demo's model is pinned to one host and its fallback to another, because the Ask box's eval passed on those hosts only (D-207, D-217). The chat client already had optional Route and Reasoning fields that only the bench set; the plugin now sets them from the entry. A pin has no fallback to other hosts, so a pinned host that is down fails the call and the router goes to the route's next model. Hosts and reasoning are OpenRouter's request members and OpenAI's own API refuses unknown ones, so another base URL is refused at start instead of silently ignored.

## Source

> The author's M4 brief: model entries with `google/gemma-4-26b-a4b-it`, host pinning `order: ["nextbit/bf16"]`, `allow_fallbacks: false`, and a fallback `openai/gpt-oss-120b` with reasoning low pinned to one host; per-model-entry host pinning and a per-entry reasoning setting in the model-openai plugin's entry config, sent only to OpenRouter (M4 prep brief from the design session, 2026-11-14).
