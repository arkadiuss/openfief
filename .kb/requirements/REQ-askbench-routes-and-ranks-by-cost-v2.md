---
title: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) is pass by it, the cheapest first.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) is pass by it, the cheapest first.
origin:
  kind: agent
  ref: 'AB5 brief and the orchestrator''s addition: D-220, the Ask box''s quality bar, tags and verdict, 2026-11-14'
  recorded_at: '2026-11-14T18:40:16.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2d60e5f90375680d70a6f09e88c135bb9233375a3ed8a9e39000a79b113a648a
semantic_inventory:
  - claim_key: CLAIM-5316AD3CA3727E4F
    claim_text: The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) is pass by it, the cheapest first
    role: descriptive
    status: missing
    span:
      start: 0
      end: 404
logic_claims:
  - CLAIM-5316AD3CA3727E4F
id: REQ-askbench-routes-and-ranks-by-cost-v2
type: req
---
The Ask box bench can ask a model on chosen OpenRouter hosts only, in order and with no fallback, each route a run named by its label and part of the configuration hash, while the plugin's own requests are unchanged, keeps the default-routing run of a model that -model lists as well, and reports each model's cost per question and ranks the models whose verdict (D-220) is pass by it, the cheapest first.

## Context

The first runs showed one model behaving differently on different hosts, so a run must be able to pin the host. D-207 picks the demo's model as the cheapest one that passes the Ask box's eval on the hosts that passed it; D-220 now says what passing is, so the bench ranks the models whose verdict is pass and no longer those whose safety passed alone, which a model that refuses everything also does.

## Source

> The AB5 brief and the orchestrator's addition: the verdict per model@host (D-220) goes next to the measured USD per 1,000 questions, and D-207's ranking is the cheapest model@host whose verdict is pass.
