---
title: The Ask box bench can set OpenRouter's reasoning control per model or label (off, low, medium or high, sent only on an OpenRouter base URL, part of the configuration hash, and sent by no other client), and a model that its -model flag lists as well as its -route entries keeps its default-routing run beside the routed ones.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench can set OpenRouter's reasoning control per model or label (off, low, medium or high, sent only on an OpenRouter base URL, part of the configuration hash, and sent by no other client), and a model that its -model flag lists as well as its -route entries keeps its default-routing run beside the routed ones.
tags:
  - askbench
origin:
  kind: agent
  ref: 'AB4 brief from the design session: softer decline wording, injection cases split into safety and answer, reasoning control and exact cost in the bench, 2026-11-14'
  recorded_at: '2026-11-14T16:54:56.910Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5777a46ed2e70492034b31e59292da8529ebfeca5baffe9fd7eac9ee31919ee8
semantic_inventory:
  - claim_key: CLAIM-A33C10D29FCD1788
    claim_text: The Ask box bench can set OpenRouter's reasoning control per model or label (off, low, medium or high, sent only on an OpenRouter base URL, part of the configuration hash, and sent by no other client), and a model that its -model flag lists as well as its -route entries keeps its default-routing run beside the routed ones
    role: descriptive
    status: missing
    span:
      start: 0
      end: 323
logic_claims:
  - CLAIM-A33C10D29FCD1788
id: REQ-askbench-reasoning-control-and-default-run
type: req
---
The Ask box bench can set OpenRouter's reasoning control per model or label (off, low, medium or high, sent only on an OpenRouter base URL, part of the configuration hash, and sent by no other client), and a model that its -model flag lists as well as its -route entries keeps its default-routing run beside the routed ones.

## Context

Reasoning models spend most of their tokens thinking, and the bench could not compare a model with its thinking off or turned down; and a model pinned to hosts lost its default-routing run, so the pinned and the unpinned could not be compared in one run (2026-11-14).

## Source

> AB4 brief, decision (c): Reasoning control per model/label: -reasoning 'model=off|low|medium|high' -> OpenRouter's reasoning request parameter; send nothing when unset; only on OpenRouter base URLs; in the config hash; an optional field in openaichat that production never sets. Fix: when a model has -route entries, also keep its default-routing run if the model id is listed in -model.
